package usecase

import (
	"context"
	"errors"
	"time"

	authdomain "github.com/afifudin23/saepedia-api/internal/auth/domain"
	userdomain "github.com/afifudin23/saepedia-api/internal/user/domain"
	"github.com/afifudin23/saepedia-api/pkg/helper"
	"github.com/afifudin23/saepedia-api/pkg/jwt"
)

type authUsecase struct {
	userRepo  userdomain.UserRepository
	wallet    authdomain.WalletReader
	revoker   authdomain.TokenRevoker
	accessKey string
}

func New(userRepo userdomain.UserRepository, wallet authdomain.WalletReader, revoker authdomain.TokenRevoker, accessKey string) authdomain.AuthUsecase {
	return &authUsecase{userRepo: userRepo, wallet: wallet, revoker: revoker, accessKey: accessKey}
}

// Logout memasukkan token ke denylist sampai waktu kadaluarsanya.
func (uc *authUsecase) Logout(ctx context.Context, jti string, exp time.Time) error {
	if jti == "" {
		return nil
	}
	return uc.revoker.Revoke(ctx, jti, exp)
}

func (uc *authUsecase) Register(ctx context.Context, in authdomain.RegisterInput) (*authdomain.AuthResult, error) {
	roles := in.Roles
	if len(roles) == 0 {
		roles = []string{"buyer"} // default role bila tidak dipilih
	}
	roles = dedupe(roles)

	hashed, err := helper.Hash(in.Password)
	if err != nil {
		return nil, err
	}

	user := &userdomain.User{
		Email:    in.Email,
		Password: hashed,
		IsAdmin:  false,
	}
	if err := uc.userRepo.Create(ctx, user, roles); err != nil {
		return nil, err
	}

	return uc.buildResult(user)
}

func (uc *authUsecase) Login(ctx context.Context, in authdomain.LoginInput) (*authdomain.AuthResult, error) {
	user, err := uc.userRepo.FindByEmail(ctx, in.Email)
	if err != nil {
		if errors.Is(err, userdomain.ErrUserNotFound) {
			return nil, authdomain.ErrInvalidCredentials
		}
		return nil, err
	}

	if !helper.Verify(in.Password, user.Password) {
		return nil, authdomain.ErrInvalidCredentials
	}

	return uc.buildResult(user)
}

func (uc *authUsecase) SelectRole(ctx context.Context, userID, role string) (*authdomain.AuthResult, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if !user.HasRole(role) {
		return nil, authdomain.ErrRoleNotOwned
	}

	token, err := jwt.Generate(user.ID, role, uc.accessKey)
	if err != nil {
		return nil, err
	}
	return &authdomain.AuthResult{
		User:       user,
		Token:      token,
		ActiveRole: role,
	}, nil
}

func (uc *authUsecase) Me(ctx context.Context, userID string) (*userdomain.User, error) {
	return uc.userRepo.FindByID(ctx, userID)
}

func (uc *authUsecase) BalanceSummary(ctx context.Context, userID string) (*authdomain.BalanceSummary, error) {
	var balance int64
	if uc.wallet != nil {
		b, err := uc.wallet.BalanceByUserID(ctx, userID)
		if err != nil {
			return nil, err
		}
		balance = b
	}
	// SellerIncome & DriverEarnings adalah placeholder (level berikutnya).
	return &authdomain.BalanceSummary{
		WalletBalance:  balance,
		SellerIncome:   0,
		DriverEarnings: 0,
	}, nil
}

// buildResult menentukan role aktif default + token setelah register/login.
// - Admin: role aktif otomatis "admin".
// - Tepat 1 role non-admin: role itu langsung aktif.
// - Lebih dari 1 role: token tetap diberi tapi role aktif kosong (harus pilih).
func (uc *authUsecase) buildResult(user *userdomain.User) (*authdomain.AuthResult, error) {
	active := ""
	need := false

	switch {
	case user.IsAdmin:
		active = "admin"
	case len(user.Roles) == 1:
		active = user.Roles[0]
	default:
		need = true
	}

	token, err := jwt.Generate(user.ID, active, uc.accessKey)
	if err != nil {
		return nil, err
	}
	return &authdomain.AuthResult{
		User:              user,
		Token:             token,
		ActiveRole:        active,
		NeedRoleSelection: need,
	}, nil
}

func dedupe(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, v := range in {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}
