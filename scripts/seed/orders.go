package main

import (
	"fmt"
	"time"

	orderdomain "github.com/afifudin23/saepedia-api/internal/order/domain"
	"gorm.io/gorm"
)

type orderSpec struct {
	BuyerEmail  string
	StoreName   string
	ProductName string
	Qty         int
	Method      string // instant / next_day / regular
	Status      string
	DriverEmail string // "" bila belum ada driver
}

// Orders mengisi 10 order demo lengkap dengan order_items, order_status_histories,
// dan wallet_transactions; stok produk & saldo wallet ikut disesuaikan agar
// konsisten. Idempotent: dilewati bila tabel orders sudah berisi data.
func Orders(db *gorm.DB) error {
	var count int64
	if err := db.Raw("SELECT COUNT(*) FROM orders").Scan(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		fmt.Println("  - orders sudah ada, dilewati")
		return nil
	}

	specs := []orderSpec{
		{"buyer1@seapedia.test", "Toko Sumber Rejeki", "Topi Baseball", 1, "regular", orderdomain.StatusSelesai, "driver1@seapedia.test"},
		{"buyer1@seapedia.test", "Elektronik Jaya", "Mouse Wireless", 1, "next_day", orderdomain.StatusSelesai, "driver2@seapedia.test"},
		{"buyer1@seapedia.test", "Dapur Sehat", "Tumbler Stainless", 2, "instant", orderdomain.StatusDikirim, "driver1@seapedia.test"},
		{"buyer1@seapedia.test", "Toko Sumber Rejeki", "Celana Chino", 1, "next_day", orderdomain.StatusSelesai, "driver1@seapedia.test"},
		{"buyer2@seapedia.test", "Dapur Sehat", "Talenan Kayu", 1, "regular", orderdomain.StatusDikemas, ""},
		{"buyer2@seapedia.test", "Toko Serba Ada", "Botol Minum 1L", 1, "regular", orderdomain.StatusMenungguKirim, ""},
		{"buyer3@seapedia.test", "Elektronik Jaya", "Keyboard Mekanik", 1, "instant", orderdomain.StatusSelesai, "driver2@seapedia.test"},
		{"buyer3@seapedia.test", "Elektronik Jaya", "Webcam HD", 1, "next_day", orderdomain.StatusMenungguKirim, ""},
		{"buyer3@seapedia.test", "Toko Serba Ada", "Payung Lipat", 2, "regular", orderdomain.StatusDikembalikan, ""},
		{"multi1@seapedia.test", "Toko Sumber Rejeki", "Topi Baseball", 1, "regular", orderdomain.StatusDikemas, ""},
	}

	for i, s := range specs {
		if err := seedOneOrder(db, s); err != nil {
			return fmt.Errorf("order #%d (%s): %w", i+1, s.ProductName, err)
		}
		fmt.Printf("  - order %s × %d → %s [%s]\n", s.ProductName, s.Qty, s.BuyerEmail, s.Status)
	}
	return nil
}

func seedOneOrder(db *gorm.DB, s orderSpec) error {
	// ── Resolusi data (read) ──────────────────────────────────────
	var buyerID string
	if err := db.Raw("SELECT id FROM users WHERE email = ?", s.BuyerEmail).Scan(&buyerID).Error; err != nil {
		return err
	}
	if buyerID == "" {
		return fmt.Errorf("buyer %s tidak ditemukan", s.BuyerEmail)
	}

	var product struct {
		ID      string
		StoreID string
		Price   int64
	}
	err := db.Raw(`SELECT p.id, p.store_id, p.price FROM products p
		JOIN stores st ON st.id = p.store_id
		WHERE st.name = ? AND p.name = ?`, s.StoreName, s.ProductName).Scan(&product).Error
	if err != nil {
		return err
	}
	if product.ID == "" {
		return fmt.Errorf("produk %q di toko %q tidak ditemukan", s.ProductName, s.StoreName)
	}

	var driverID *string
	if s.DriverEmail != "" {
		var id string
		if err := db.Raw("SELECT id FROM users WHERE email = ?", s.DriverEmail).Scan(&id).Error; err != nil {
			return err
		}
		if id != "" {
			driverID = &id
		}
	}

	// Alamat: pakai alamat utama buyer bila ada, fallback dummy.
	var addr struct {
		RecipientName string
		Phone         string
		FullAddress   string
	}
	db.Raw(`SELECT recipient_name, phone, full_address FROM addresses
		WHERE user_id = ? ORDER BY is_primary DESC, created_at ASC LIMIT 1`, buyerID).Scan(&addr)
	if addr.RecipientName == "" {
		addr.RecipientName = s.BuyerEmail
		addr.Phone = "0812000000"
		addr.FullAddress = "Alamat demo"
	}

	// ── Hitung biaya (pakai aturan domain) ────────────────────────
	fee, ok := orderdomain.DeliveryFee(s.Method)
	if !ok {
		return fmt.Errorf("metode kirim %q invalid", s.Method)
	}
	subtotal := product.Price * int64(s.Qty)
	tax := orderdomain.CalcTax(subtotal) // discount 0
	total := subtotal + fee + tax

	now := time.Now()
	var takenAt, completedAt, refundedAt *time.Time
	var earning int64
	switch s.Status {
	case orderdomain.StatusDikirim:
		takenAt = &now
	case orderdomain.StatusSelesai:
		takenAt, completedAt = &now, &now
		earning = orderdomain.CalcDriverEarning(fee)
	case orderdomain.StatusDikembalikan:
		refundedAt = &now
	}

	// ── Eksekusi transaksional ────────────────────────────────────
	return db.Transaction(func(tx *gorm.DB) error {
		refunded := s.Status == orderdomain.StatusDikembalikan

		// 1. Stok: order normal mengurangi stok; order dikembalikan tidak (net 0).
		if !refunded {
			res := tx.Exec("UPDATE products SET stock = stock - ?, updated_at = NOW() WHERE id = ? AND stock >= ?",
				s.Qty, product.ID, s.Qty)
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return fmt.Errorf("stok %q tidak cukup", s.ProductName)
			}
		}

		// 2. Buat order.
		var orderID string
		err := tx.Raw(`INSERT INTO orders
			(buyer_id, store_id, recipient_name, phone, full_address, delivery_method,
			 subtotal, discount, discount_code, delivery_fee, tax, total, status,
			 driver_id, driver_earning, taken_at, completed_at, refunded_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, 0, '', ?, ?, ?, ?, ?, ?, ?, ?, ?)
			RETURNING id`,
			buyerID, product.StoreID, addr.RecipientName, addr.Phone, addr.FullAddress, s.Method,
			subtotal, fee, tax, total, s.Status,
			driverID, earning, takenAt, completedAt, refundedAt,
		).Scan(&orderID).Error
		if err != nil {
			return err
		}

		// 3. Order item.
		if err := tx.Exec(`INSERT INTO order_items
			(order_id, product_id, product_name, price, quantity, subtotal)
			VALUES (?, ?, ?, ?, ?, ?)`,
			orderID, product.ID, s.ProductName, product.Price, s.Qty, subtotal,
		).Error; err != nil {
			return err
		}

		// 4. Riwayat status.
		for _, h := range historyFor(s.Status) {
			if err := tx.Exec("INSERT INTO order_status_histories (order_id, status, note) VALUES (?, ?, ?)",
				orderID, h.status, h.note).Error; err != nil {
				return err
			}
		}

		// 5. Wallet: pastikan ada, lalu catat transaksi.
		if err := tx.Exec("INSERT INTO wallets (user_id) VALUES (?) ON CONFLICT (user_id) DO NOTHING", buyerID).Error; err != nil {
			return err
		}
		var walletID string
		var balance int64
		if err := tx.Raw("SELECT id, balance FROM wallets WHERE user_id = ? FOR UPDATE", buyerID).Row().Scan(&walletID, &balance); err != nil {
			return err
		}

		afterPayment := balance - total
		if err := tx.Exec(`INSERT INTO wallet_transactions (wallet_id, type, amount, balance_after, description)
			VALUES (?, 'payment', ?, ?, ?)`, walletID, -total, afterPayment, "payment for order "+orderID).Error; err != nil {
			return err
		}

		if refunded {
			// Refund mengembalikan dana → saldo bersih tidak berubah.
			if err := tx.Exec(`INSERT INTO wallet_transactions (wallet_id, type, amount, balance_after, description)
				VALUES (?, 'refund', ?, ?, ?)`, walletID, total, afterPayment+total, "auto refund for overdue order "+orderID).Error; err != nil {
				return err
			}
			return nil // balance tetap
		}

		// Order normal: saldo berkurang.
		return tx.Exec("UPDATE wallets SET balance = ?, updated_at = NOW() WHERE id = ?", afterPayment, walletID).Error
	})
}

type statusNote struct{ status, note string }

func historyFor(status string) []statusNote {
	dikemas := statusNote{orderdomain.StatusDikemas, "order created after successful checkout"}
	menunggu := statusNote{orderdomain.StatusMenungguKirim, "order processed by seller, ready for pickup"}
	dikirim := statusNote{orderdomain.StatusDikirim, "driver took the job, package on the way"}
	selesai := statusNote{orderdomain.StatusSelesai, "driver confirmed delivery completed"}
	dikembalikan := statusNote{orderdomain.StatusDikembalikan, "auto refund: order overdue (delivery SLA exceeded)"}

	switch status {
	case orderdomain.StatusMenungguKirim:
		return []statusNote{dikemas, menunggu}
	case orderdomain.StatusDikirim:
		return []statusNote{dikemas, menunggu, dikirim}
	case orderdomain.StatusSelesai:
		return []statusNote{dikemas, menunggu, dikirim, selesai}
	case orderdomain.StatusDikembalikan:
		return []statusNote{dikemas, menunggu, dikembalikan}
	default:
		return []statusNote{dikemas}
	}
}
