package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"golang.org/x/term"
)

// ── Config ────────────────────────────────────────────────────

const (
	projectName = "SEAPEDIA API"
	prodBranch  = "main"
	devBranch   = "dev"
)

var projectRoot string

func init() {
	wd, _ := os.Getwd()
	projectRoot = wd
}

// ── ANSI Colors ───────────────────────────────────────────────

const (
	GR = "\033[92m"
	YL = "\033[93m"
	CY = "\033[96m"
	RD = "\033[91m"
	MG = "\033[95m"
	WH = "\033[97m"
	DM = "\033[2m"
	BD = "\033[1m"
	RS = "\033[0m"
)

// ── Terminal Input ────────────────────────────────────────────

type keyEvent struct {
	isArrow bool
	key     byte
}

func readKey() keyEvent {
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return keyEvent{key: '\r'}
	}
	defer term.Restore(fd, oldState)

	buf := make([]byte, 1)
	if _, err := os.Stdin.Read(buf); err != nil {
		return keyEvent{key: '\r'}
	}
	ch := buf[0]

	if ch == 0x03 {
		fmt.Printf("\n%s❌ Dibatalkan.%s\n\n", RD, RS)
		os.Exit(0)
	}

	// Windows native: arrow keys dimulai 0xe0 atau 0x00
	if runtime.GOOS == "windows" && (ch == 0xe0 || ch == 0x00) {
		extra := make([]byte, 1)
		os.Stdin.Read(extra)
		switch extra[0] {
		case 0x48:
			return keyEvent{isArrow: true, key: 'H'}
		case 0x50:
			return keyEvent{isArrow: true, key: 'P'}
		}
		return keyEvent{key: ch}
	}

	// Unix/WSL: escape sequence \x1b[A / \x1b[B
	if ch == 0x1b {
		seq := make([]byte, 2)
		n, _ := os.Stdin.Read(seq)
		if n >= 2 && seq[0] == '[' {
			switch seq[1] {
			case 'A':
				return keyEvent{isArrow: true, key: 'H'}
			case 'B':
				return keyEvent{isArrow: true, key: 'P'}
			}
		}
		return keyEvent{key: ch}
	}

	return keyEvent{key: ch}
}

func arrowUI(title string, options []string, defaultSel int) int {
	sel := defaultSel
	n := len(options)

	draw := func() {
		for i, opt := range options {
			if i == sel {
				fmt.Printf("  %s%s❯ %s%s\n", GR, BD, opt, RS)
			} else {
				fmt.Printf("  %s  %s%s\n", DM, opt, RS)
			}
		}
	}

	fmt.Printf("\n  %s%s%s\n", BD, title, RS)
	draw()

	for {
		ev := readKey()

		if ev.isArrow {
			if ev.key == 'H' {
				sel = (sel - 1 + n) % n
			} else if ev.key == 'P' {
				sel = (sel + 1) % n
			} else {
				continue
			}
		} else if ev.key == '\r' || ev.key == '\n' {
			break
		} else {
			continue
		}

		fmt.Printf("\033[%dA", n)
		draw()
	}

	// Final static render
	fmt.Printf("\033[%dA", n)
	for range options {
		fmt.Print("\r\033[2K\n")
	}
	fmt.Printf("\033[%dA", n)
	fmt.Printf("  %s%s❯ %s%s\n", GR, BD, options[sel], RS)
	return sel
}

func confirm(question string, defaultYes bool) bool {
	opts := []string{GR + "Ya" + RS, RD + "Tidak" + RS}
	def := 0
	if !defaultYes {
		def = 1
	}
	return arrowUI(question, opts, def) == 0
}

func menuSelect(title string, options []string) int {
	return arrowUI(title, options, 0)
}

// ── Helpers ───────────────────────────────────────────────────

func runCmd(cmdStr string, show bool) (string, int) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("powershell", "-NoProfile", "-Command", cmdStr)
	} else {
		cmd = exec.Command("sh", "-c", cmdStr)
	}
	cmd.Dir = projectRoot

	if show {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				return "", exitErr.ExitCode()
			}
			return "", 1
		}
		return "", 0
	}

	out, err := cmd.Output()
	result := strings.TrimSpace(string(out))
	if err != nil {
		return result, 1
	}
	return result, 0
}

func run(cmdStr string) string {
	out, _ := runCmd(cmdStr, false)
	return out
}

func runShow(cmdStr string) int {
	_, code := runCmd(cmdStr, true)
	return code
}

func gitCommit(subject, body string, allowEmpty bool) {
	msg := subject
	if body != "" {
		msg = subject + "\n\n" + body
	}
	f, err := os.CreateTemp("", "git-commit-*.txt")
	if err != nil {
		return
	}
	defer os.Remove(f.Name())
	f.WriteString(msg)
	f.Close()

	args := []string{"commit", "-F", f.Name()}
	if allowEmpty {
		args = append(args, "--allow-empty")
	}
	cmd := exec.Command("git", args...)
	cmd.Dir = projectRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func ask(question string) string {
	fmt.Printf("  %s❯%s %s", CY, RS, question)
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text())
	}
	fmt.Printf("\n%s❌ Dibatalkan.%s\n", RD, RS)
	os.Exit(0)
	return ""
}

func sep(color string) {
	fmt.Printf("%s%s%s\n", color, strings.Repeat("─", 60), RS)
}

func clearScreen() {
	fmt.Print("\033[2J\033[H")
}

func pauseBack() {
	fmt.Printf("\n  %sTekan Enter untuk kembali ke menu...%s", DM, RS)
	bufio.NewReader(os.Stdin).ReadString('\n')
}

// ── Git Info ──────────────────────────────────────────────────

func getCurrentBranch() string {
	return run("git rev-parse --abbrev-ref HEAD")
}

func getTagsMatching(pattern string) []string {
	out := run(fmt.Sprintf("git tag -l %q --sort=-v:refname", pattern))
	if out == "" {
		return nil
	}
	var tags []string
	for _, t := range strings.Split(out, "\n") {
		if t != "" {
			tags = append(tags, t)
		}
	}
	return tags
}

func getTagMessage(tag string) string {
	out := run(fmt.Sprintf("git tag -l -n1 %q", tag))
	return strings.TrimSpace(strings.TrimPrefix(out, tag))
}

func getTagDate(tag string) string {
	d := run(fmt.Sprintf("git log -1 --format=%%ci %q", tag))
	if len(d) >= 16 {
		return d[:16]
	}
	return "-"
}

func getLatestTagOnBranch(branch string) string {
	result := run(fmt.Sprintf("git describe --tags --abbrev=0 %s", branch))
	if result == "" {
		result = run(fmt.Sprintf("git describe --tags --abbrev=0 origin/%s", branch))
	}
	if result == "" {
		return "-"
	}
	return result
}

type commitInfo struct {
	sha  string
	msg  string
	date string
}

func getLatestCommit(branch string) commitInfo {
	sha := run(fmt.Sprintf("git log -1 --format=%%h %s", branch))
	if sha == "" {
		sha = run(fmt.Sprintf("git log -1 --format=%%h origin/%s", branch))
	}
	msg := run(fmt.Sprintf("git log -1 --format=%%s %s", branch))
	if msg == "" {
		msg = run(fmt.Sprintf("git log -1 --format=%%s origin/%s", branch))
	}
	date := run(fmt.Sprintf("git log -1 --format=%%ci %s", branch))
	if date == "" {
		date = run(fmt.Sprintf("git log -1 --format=%%ci origin/%s", branch))
	}
	if sha == "" {
		sha = "-"
	}
	if msg == "" {
		msg = "-"
	}
	if len(msg) > 60 {
		msg = msg[:60]
	}
	if date == "" {
		date = "-"
	} else if len(date) >= 16 {
		date = date[:16]
	}
	return commitInfo{sha: sha, msg: msg, date: date}
}

// ── Version ───────────────────────────────────────────────────

func getVersion() string {
	data, err := os.ReadFile(filepath.Join(projectRoot, "config", "version.go"))
	if err != nil {
		return "unknown"
	}
	re := regexp.MustCompile(`Version\s*=\s*"(\d+\.\d+\.\d+)"`)
	m := re.FindStringSubmatch(string(data))
	if m == nil {
		return "unknown"
	}
	return m[1]
}

func setVersion(newVersion string) {
	path := filepath.Join(projectRoot, "config", "version.go")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	re := regexp.MustCompile(`(Version\s*=\s*)"[^"]*"`)
	updated := re.ReplaceAllString(string(data), `${1}"`+newVersion+`"`)
	os.WriteFile(path, []byte(updated), 0644)
}

func bumpVersion(version, bumpType string) string {
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return version
	}
	major, _ := strconv.Atoi(parts[0])
	minor, _ := strconv.Atoi(parts[1])
	patch, _ := strconv.Atoi(parts[2])

	switch bumpType {
	case "major":
		major++
		minor = 0
		patch = 0
	case "minor":
		minor++
		patch = 0
	case "patch":
		patch++
	}
	return fmt.Sprintf("%d.%d.%d", major, minor, patch)
}

func getChangelogDescription(version string) string {
	data, err := os.ReadFile(filepath.Join(projectRoot, "CHANGELOG.md"))
	if err != nil {
		return ""
	}
	content := string(data)
	header := fmt.Sprintf("## [%s]", version)
	start := strings.Index(content, header)
	if start == -1 {
		return ""
	}
	end := strings.Index(content[start:], "\n---")
	if end == -1 {
		return strings.TrimSpace(content[start:])
	}
	return strings.TrimSpace(content[start : start+end])
}

// ── Remote Connectivity ───────────────────────────────────────

func getRemoteURL() string {
	return run("git remote get-url origin")
}

func parseRemoteProto(url string) string {
	if strings.HasPrefix(url, "git@") || strings.HasPrefix(url, "ssh://") {
		return "SSH"
	}
	if strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://") {
		return "HTTPS"
	}
	return "unknown"
}

func maskURL(url string) string {
	re := regexp.MustCompile(`(https?://)([^@]+@)`)
	return re.ReplaceAllString(url, "$1***@")
}

func checkRemoteConnectivity() bool {
	url := getRemoteURL()
	if url == "" {
		fmt.Printf("   %s❌ Remote 'origin' tidak ditemukan.%s\n", RD, RS)
		return false
	}
	proto := parseRemoteProto(url)
	fmt.Printf("   %s🔗 Remote (%s): %s%s\n", DM, proto, maskURL(url), RS)
	fmt.Printf("   %s⏳ Mengecek koneksi ke remote...%s\n", DM, RS)

	if proto == "SSH" {
		cmd := exec.Command("ssh", "-T", "-o", "StrictHostKeyChecking=no", "-o", "ConnectTimeout=10", "git@github.com")
		out, _ := cmd.CombinedOutput()
		output := strings.TrimSpace(string(out))
		if strings.Contains(output, "successfully authenticated") || strings.Contains(output, "Hi ") {
			fmt.Printf("   %s✅ SSH OK — %s%s\n", GR, output, RS)
			return true
		}
		if len(output) > 150 {
			output = output[:150]
		}
		fmt.Printf("   %s❌ SSH gagal:%s %s\n", RD, RS, output)
		return false
	}

	if proto == "HTTPS" {
		cmd := exec.Command("git", "ls-remote", "--exit-code", "--heads", "origin")
		cmd.Dir = projectRoot
		cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
		err := cmd.Run()
		if err == nil {
			fmt.Printf("   %s✅ HTTPS OK — autentikasi berhasil.%s\n", GR, RS)
			return true
		}
		fmt.Printf("   %s❌ HTTPS gagal%s\n", RD, RS)
		return false
	}

	fmt.Printf("   %s⚠️  Format remote URL tidak dikenali: %s%s\n", YL, url, RS)
	return false
}

func switchRemoteProtocol() bool {
	url := getRemoteURL()
	proto := parseRemoteProto(url)
	var newURL, newProto string

	if proto == "SSH" {
		path := strings.TrimSuffix(strings.SplitN(url, ":", 2)[1], ".git")
		newURL = "https://github.com/" + path + ".git"
		newProto = "HTTPS"
	} else if proto == "HTTPS" {
		re := regexp.MustCompile(`https?://[^/]+/`)
		path := strings.TrimSuffix(re.ReplaceAllString(url, ""), ".git")
		newURL = "git@github.com:" + path + ".git"
		newProto = "SSH"
	} else {
		fmt.Printf("   %s⚠️  Tidak bisa konversi URL: %s%s\n", YL, url, RS)
		return false
	}

	run(fmt.Sprintf("git remote set-url origin %s", newURL))
	fmt.Printf("   %s✅ Remote diganti ke %s: %s%s\n", GR, newProto, maskURL(newURL), RS)
	return true
}

func checkAndConfirmRemote() bool {
	fmt.Println("")
	sep(DM)
	fmt.Printf("%s🔌  Cek Koneksi Remote%s\n", BD, RS)
	sep(DM)

	if checkRemoteConnectivity() {
		return true
	}

	url := getRemoteURL()
	proto := parseRemoteProto(url)
	alt := "HTTPS"
	if proto == "SSH" {
		alt = "SSH"
	} else {
		alt = "SSH"
	}

	fmt.Println("")
	opts := []string{
		RD + "🚫  Keluar dari Release Manager" + RS,
		GR + "⚡  Ganti ke " + alt + " lalu coba lagi" + RS,
		DM + "↩️   Kembali ke menu utama" + RS,
	}
	idx := menuSelect(fmt.Sprintf("Koneksi %s gagal. Pilih aksi:", proto), opts)

	if idx == 0 {
		fmt.Printf("\n%s👋 Bye!%s\n\n", GR, RS)
		os.Exit(0)
	}
	if idx == 2 {
		return false
	}

	switchRemoteProtocol()
	fmt.Printf("   %s⏳ Cek ulang koneksi...%s\n", DM, RS)
	if !checkRemoteConnectivity() {
		fmt.Printf("   %s❌ Koneksi masih gagal setelah ganti protokol.%s\n", RD, RS)
		fmt.Printf("\n  %sTekan Enter untuk keluar...%s", DM, RS)
		bufio.NewReader(os.Stdin).ReadString('\n')
		return false
	}
	return true
}

func offerCheckoutDev(currentBranch string) {
	fmt.Printf("   %s⚠️  Release penuh hanya tersedia di branch '%s'.%s\n", YL, devBranch, RS)
	if confirm(fmt.Sprintf("🔀  Pindah ke branch '%s' sekarang?", devBranch), true) {
		cmd := exec.Command("git", "checkout", devBranch)
		cmd.Dir = projectRoot
		out, err := cmd.CombinedOutput()
		if err == nil {
			fmt.Printf("   %s✅ Pindah ke '%s'. Jalankan release lagi dari menu.%s\n", GR, devBranch, RS)
		} else {
			sep(RD)
			fmt.Printf("%s❌  Gagal pindah ke branch '%s':%s\n", RD, devBranch, RS)
			fmt.Printf("   %s%s%s\n", DM, strings.TrimSpace(string(out)), RS)
			sep(RD)
			fmt.Printf("\n  %sTekan Enter untuk kembali ke menu  |  q + Enter untuk keluar: %s", DM, RS)
			sc := bufio.NewScanner(os.Stdin)
			if sc.Scan() && strings.ToLower(strings.TrimSpace(sc.Text())) == "q" {
				fmt.Printf("\n%s👋 Bye!%s\n\n", GR, RS)
				os.Exit(0)
			}
		}
	}
}

// ── Status ────────────────────────────────────────────────────

func showStatus() {
	fmt.Printf("\n%s%s🚀  %s — Status%s\n", CY, BD, projectName, RS)
	sep(DM)
	fmt.Printf("📁  Branch saat ini  : %s%s%s\n", CY, getCurrentBranch(), RS)
	fmt.Printf("📦  Versi            : %sv%s%s\n", YL, getVersion(), RS)
	sep(DM)

	fmt.Printf("\n%s📊  Status Per Branch%s\n", BD, RS)
	sep(DM)

	for _, b := range []struct {
		icon, branch string
	}{{"🔧", devBranch}, {"🟢", prodBranch}} {
		tag := getLatestTagOnBranch(b.branch)
		c := getLatestCommit(b.branch)
		fmt.Printf("%s  %s%-10s%s → Tag: %s%-18s%s Commit: %s%s (%s)%s\n",
			b.icon, BD, b.branch, RS, GR, tag, RS, DM, c.sha, c.date, RS)
		fmt.Printf("              → %s%s%s\n", DM, c.msg, RS)
	}

	current := getCurrentBranch()
	if current != devBranch && current != prodBranch {
		tag := getLatestTagOnBranch(current)
		c := getLatestCommit(current)
		fmt.Printf("🌿  %s%-10s%s → Tag: %s%-18s%s Commit: %s%s (%s)%s\n",
			BD, current, RS, GR, tag, RS, DM, c.sha, c.date, RS)
		fmt.Printf("              → %s%s%s\n", DM, c.msg, RS)
	}
	sep(DM)
}

func showTags() {
	fmt.Printf("\n%s🏷️   Tag Releases%s\n", BD, RS)
	sep(DM)
	tags := getTagsMatching("v*")
	if len(tags) == 0 {
		fmt.Printf("   %sBelum ada tag.%s\n", DM, RS)
	} else {
		fmt.Printf("   Total: %s%d tag%s\n\n", YL, len(tags), RS)
		limit := len(tags)
		if limit > 20 {
			limit = 20
		}
		for _, t := range tags[:limit] {
			fmt.Printf("   %s🟢 %-22s%s %s%s  %s%s\n", GR, t, RS, DM, getTagDate(t), getTagMessage(t), RS)
		}
		if len(tags) > 20 {
			fmt.Printf("\n   %s... dan %d tag lainnya%s\n", DM, len(tags)-20, RS)
		}
	}
	sep(DM)
}

// ── Release ───────────────────────────────────────────────────

func release() {
	fmt.Printf("\n%s%s🚀  %s — Release%s\n", CY, BD, projectName, RS)
	sep(DM)

	fmt.Printf("%s⏳ Fetching dari remote...%s\n", DM, RS)
	run("git fetch --all --tags --force")

	currentVersion := getVersion()
	currentBranch := getCurrentBranch()

	fmt.Printf("\n📁  Branch  : %s%s%s\n", CY, currentBranch, RS)
	fmt.Printf("📦  Versi   : %sv%s%s\n", YL, currentVersion, RS)
	sep(DM)

	fmt.Printf("  %s📋  Rules:%s\n", BD, RS)
	fmt.Printf("  %s• current       → Release penuh saja, wajib di branch '%s'%s\n", WH, devBranch, RS)
	fmt.Printf("  %s• CHANGELOG.md harus sudah punya entry versi target sebelum release%s\n", WH, RS)
	fmt.Printf("  %s• Tag versi tidak boleh sudah ada di remote sebelumnya%s\n", WH, RS)
	fmt.Printf("  %s• Sync          → Auto pull --rebase, conflict = proses dibatalkan%s\n", WH, RS)

	bumpOpts := []string{
		fmt.Sprintf("current →  %sv%s%s  (Release penuh, tanpa bump)", CY, currentVersion, RS),
		fmt.Sprintf("patch   →  %sv%s%s", WH, bumpVersion(currentVersion, "patch"), RS),
		fmt.Sprintf("minor   →  %sv%s%s", YL, bumpVersion(currentVersion, "minor"), RS),
		fmt.Sprintf("major   →  %sv%s%s", MG, bumpVersion(currentVersion, "major"), RS),
		fmt.Sprintf("custom  →  %smasukkan versi manual%s", DM, RS),
		fmt.Sprintf("%s↩️   Kembali ke menu utama%s", DM, RS),
	}

	bumpIdx := menuSelect("Pilih tipe bump:", bumpOpts)
	if bumpIdx == 5 {
		return
	}

	var newVersion string
	switch bumpIdx {
	case 0:
		newVersion = currentVersion
	case 1:
		newVersion = bumpVersion(currentVersion, "patch")
	case 2:
		newVersion = bumpVersion(currentVersion, "minor")
	case 3:
		newVersion = bumpVersion(currentVersion, "major")
	default:
		newVersion = ask("Masukkan versi baru (contoh: 1.0.0): ")
		matched, _ := regexp.MatchString(`^\d+\.\d+\.\d+$`, newVersion)
		if !matched {
			fmt.Printf("%s❌ Format versi tidak valid. Gunakan X.Y.Z%s\n", RD, RS)
			pauseBack()
			return
		}
	}

	changelogDesc := getChangelogDescription(newVersion)
	if changelogDesc == "" {
		fmt.Printf("\n%s❌ CHANGELOG.md belum diupdate untuk v%s!%s\n", RD, newVersion, RS)
		fmt.Printf("%sTambahkan entry ## [%s] di CHANGELOG.md dulu, lalu jalankan ulang.%s\n", YL, newVersion, RS)
		pauseBack()
		return
	}

	isBumpOnly := false
	if bumpIdx == 0 {
		if currentBranch != devBranch {
			offerCheckoutDev(currentBranch)
			return
		}
	} else {
		modeIdx := menuSelect("Pilih mode:", []string{
			fmt.Sprintf("%sHanya bump%s  →  commit + push", CY, RS),
			fmt.Sprintf("%sRelease penuh%s  →  commit + tag + push", GR, RS),
			fmt.Sprintf("%s↩️   Kembali ke menu utama%s", DM, RS),
		})
		if modeIdx == 2 {
			return
		}
		isBumpOnly = modeIdx == 0
		if !isBumpOnly && currentBranch != devBranch {
			offerCheckoutDev(currentBranch)
			return
		}
	}

	// Sync check
	fmt.Printf("%s⏳ Mengecek sinkronisasi dengan remote (origin/%s)...%s\n", DM, devBranch, RS)
	run("git fetch origin")
	behindStr := run(fmt.Sprintf("git rev-list --count HEAD..origin/%s", devBranch))
	behind, _ := strconv.Atoi(behindStr)
	if behind > 0 {
		fmt.Printf("   %s⚠️  Tertinggal %d commit dari origin/%s — menjalankan rebase...%s\n", YL, behind, devBranch, RS)
		cmd := exec.Command("git", "rebase", fmt.Sprintf("origin/%s", devBranch))
		cmd.Dir = projectRoot
		if err := cmd.Run(); err != nil {
			exec.Command("git", "rebase", "--abort").Run()
			sep(RD)
			fmt.Printf("%s%s❌  Rebase conflict! Selesaikan dulu sebelum lanjut.%s\n", RD, BD, RS)
			sep(RD)
			pauseBack()
			return
		}
		fmt.Printf("   %s✅ Rebase terhadap origin/%s berhasil.%s\n", GR, devBranch, RS)
	} else {
		fmt.Printf("   %s✅ Branch '%s' sudah up-to-date dengan origin/%s.%s\n", GR, currentBranch, devBranch, RS)
	}

	if !checkAndConfirmRemote() {
		return
	}

	var confirmText string
	if bumpIdx == 0 {
		confirmText = fmt.Sprintf("🚀  Mau lanjut Release penuh v%s (tetap versi sekarang)?", currentVersion)
	} else {
		modeLabel := "Hanya bump version"
		if !isBumpOnly {
			modeLabel = "Release penuh"
		}
		confirmText = fmt.Sprintf("🚀  Mau lanjut %s (v%s → v%s)?", modeLabel, currentVersion, newVersion)
	}
	if !confirm(confirmText, true) {
		fmt.Printf("%s❌ Dibatalkan.%s\n", RD, RS)
		return
	}

	// ── BUMP ONLY ────────────────────────────────────────────
	if isBumpOnly {
		fmt.Printf("\n%s📋  Langkah yang akan dijalankan:%s\n", BD, RS)
		fmt.Printf("   %s1. Update VERSION → %s%s\n", DM, newVersion, RS)
		fmt.Printf("   2. git commit — chore(bump): v%s\n", newVersion)
		fmt.Printf("   %s3. git push origin %s%s\n\n", GR, currentBranch, RS)

		if newVersion != currentVersion {
			setVersion(newVersion)
			fmt.Printf("   %s✅ v%s → v%s%s\n", GR, currentVersion, newVersion, RS)
		}

		run("git add -A")
		if run("git status --porcelain") != "" {
			gitCommit(fmt.Sprintf("chore(bump): v%s", newVersion), changelogDesc, false)
			fmt.Printf("   %s✅ Committed.%s\n", GR, RS)
		} else {
			fmt.Printf("   %sℹ️  No changes to commit.%s\n", DM, RS)
		}

		pushRC := runShow(fmt.Sprintf("git push origin %s", currentBranch))
		pushOk := pushRC == 0

		fmt.Println("")
		sep(DM)
		fmt.Printf("📦  Versi lama  : %sv%s%s\n", DM, currentVersion, RS)
		fmt.Printf("🆕  Versi baru  : %s%sv%s%s\n", GR, BD, newVersion, RS)
		fmt.Printf("🌿  Branch      : %s%s%s\n", CY, currentBranch, RS)
		if pushOk {
			fmt.Printf("☁️   Push        : %s✅ berhasil%s\n", GR, RS)
			sep(GR)
			fmt.Printf("%s%s✅  Bump selesai!%s\n", GR, BD, RS)
		} else {
			fmt.Printf("☁️   Push        : %s❌ gagal — push manual:%s\n", RD, RS)
			fmt.Printf("   %s   git push origin %s%s\n", WH, currentBranch, RS)
			sep(YL)
			fmt.Printf("%s%s⚠️   Bump selesai (push gagal)!%s\n", YL, BD, RS)
		}
		sep(DM)
		os.Exit(0)
	}

	// ── RELEASE PENUH ────────────────────────────────────────
	releaseTag := fmt.Sprintf("v%s", newVersion)
	if existing := getTagsMatching(releaseTag); len(existing) > 0 {
		fmt.Printf("\n%s❌ Tag %s sudah ada! Pilih versi yang berbeda.%s\n", RD, releaseTag, RS)
		pauseBack()
		return
	}

	fmt.Printf("\n%s📋  Langkah yang akan dijalankan:%s\n", BD, RS)
	fmt.Printf("   %s1. Update VERSION → %s%s\n", DM, newVersion, RS)
	fmt.Printf("   2. git commit — chore(release): v%s\n", newVersion)
	fmt.Printf("   %s3. git tag -a %s%s\n", GR, releaseTag, RS)
	fmt.Printf("   %s4. git push origin %s + %s%s\n\n", GR, currentBranch, releaseTag, RS)

	committed := false
	rollback := func(reason string) {
		fmt.Printf("\n   %s❌ %s%s\n", RD, reason, RS)
		fmt.Printf("   %s↩️  Rolling back...%s\n", YL, RS)
		setVersion(currentVersion)
		if committed {
			run("git reset HEAD~1")
		}
		fmt.Printf("   %s↩️  Reverted ke v%s.%s\n", YL, currentVersion, RS)
	}

	// Step 1: Update version
	fmt.Printf("\n%s[1/4] Updating version...%s\n", BD, RS)
	setVersion(newVersion)
	fmt.Printf("   %s✅ v%s → v%s%s\n", GR, currentVersion, newVersion, RS)

	// Step 2: Commit
	fmt.Printf("\n%s[2/4] Committing...%s\n", BD, RS)
	run("git add -A")
	gitCommit(fmt.Sprintf("chore(release): v%s", newVersion), changelogDesc, true)
	committed = true
	fmt.Printf("   %s✅ Committed.%s\n", GR, RS)

	// Step 3: Tag
	fmt.Printf("\n%s[3/4] Creating tag...%s\n", BD, RS)
	run(fmt.Sprintf("git tag -a %s -m \"Release v%s\"", releaseTag, newVersion))
	fmt.Printf("   %s✅ Tag %s dibuat.%s\n", GR, releaseTag, RS)

	// Step 4: Push
	fmt.Printf("\n%s[4/4] Pushing ke remote...%s\n", BD, RS)
	pushBranchRC := runShow(fmt.Sprintf("git push origin %s", currentBranch))
	pushTagRC := runShow(fmt.Sprintf("git push origin %s", releaseTag))
	pushOk := pushBranchRC == 0 && pushTagRC == 0

	if !pushOk {
		rollback("Push gagal!")
		return
	}

	fmt.Println("")
	sep(DM)
	fmt.Printf("📦  Versi lama  : %sv%s%s\n", DM, currentVersion, RS)
	fmt.Printf("🆕  Versi baru  : %s%sv%s%s\n", GR, BD, newVersion, RS)
	fmt.Printf("🌿  Branch      : %s%s%s\n", CY, currentBranch, RS)
	fmt.Printf("🏷️   Tag         : %s%s%s\n", YL, releaseTag, RS)
	fmt.Printf("☁️   Push        : %s✅ berhasil%s\n", GR, RS)
	sep(GR)
	fmt.Printf("%s%s✅  Release selesai!%s\n", GR, BD, RS)
	sep(GR)
	os.Exit(0)
}

// ── Delete Tag ────────────────────────────────────────────────

func deleteTag() {
	fmt.Printf("\n%s%s🗑️   HAPUS TAG%s\n", RD, BD, RS)
	sep(DM)

	fmt.Printf("%s⏳ Fetching tags dari remote...%s\n", DM, RS)
	run("git fetch --tags --force")

	allTags := getTagsMatching("v*")
	if len(allTags) == 0 {
		fmt.Printf("\n   %sTidak ada tag untuk dihapus.%s\n", DM, RS)
		return
	}

	fmt.Printf("\n   %sDaftar tag%s %s(%d total):%s\n\n", BD, RS, DM, len(allTags), RS)
	for i, tag := range allTags {
		fmt.Printf("   %s%3d.%s %s🟢 %-22s%s %s%s  %s%s\n",
			DM, i+1, RS, GR, tag, RS, DM, getTagDate(tag), getTagMessage(tag), RS)
	}

	fmt.Printf("\n   %s💡 Ketik nomor tag yang ingin dihapus.\n", DM)
	fmt.Printf("   💡 Pisahkan dengan koma untuk hapus beberapa. Contoh: 1,2,3%s\n\n", RS)

	userInput := ask("🗑️  Pilih tag (nomor/batal): ")
	if userInput == "" || strings.ToLower(userInput) == "batal" || strings.ToLower(userInput) == "cancel" {
		fmt.Printf("%s❌ Dibatalkan.%s\n", RD, RS)
		return
	}

	var tagsToDelete []string
	for _, part := range strings.Split(userInput, ",") {
		n, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || n < 1 || n > len(allTags) {
			fmt.Printf("\n%s⚠️  Nomor %s tidak valid (harus 1-%d)%s\n", YL, part, len(allTags), RS)
			return
		}
		tagsToDelete = append(tagsToDelete, allTags[n-1])
	}

	fmt.Println("")
	sep(YL)
	fmt.Printf("%s⚠️  Tag yang akan DIHAPUS (%d):%s\n", YL, len(tagsToDelete), RS)
	for _, t := range tagsToDelete {
		fmt.Printf("   %s🟢 %s%s  —  %s%s%s\n", GR, t, RS, DM, getTagMessage(t), RS)
	}
	sep(YL)

	scopeIdx := menuSelect("Hapus dari mana?", []string{
		CY + "Lokal saja" + RS,
		YL + "Remote saja" + RS,
		RD + "Keduanya (lokal + remote)" + RS,
		DM + "↩️   Kembali ke menu utama" + RS,
	})
	if scopeIdx == 3 {
		return
	}
	scopeLabels := []string{"LOKAL", "REMOTE", "LOKAL + REMOTE"}
	scopeLabel := scopeLabels[scopeIdx]

	if !confirm(fmt.Sprintf("🗑️  Hapus %d tag dari %s?", len(tagsToDelete), scopeLabel), false) {
		fmt.Printf("%s❌ Dibatalkan.%s\n", RD, RS)
		return
	}

	fmt.Printf("\n%s⏳ Menghapus tag...%s\n\n", DM, RS)
	for _, tag := range tagsToDelete {
		fmt.Printf("   🗑️  %s%s%s...\n", YL, tag, RS)
		if scopeIdx == 1 || scopeIdx == 2 {
			r := run(fmt.Sprintf("git push origin --delete %s", tag))
			if strings.Contains(strings.ToLower(r), "error") || strings.Contains(strings.ToLower(r), "fatal") {
				fmt.Printf("      %s❌ Remote gagal%s\n", RD, RS)
			} else {
				fmt.Printf("      %s✅ Remote dihapus%s\n", GR, RS)
			}
		}
		if scopeIdx == 0 || scopeIdx == 2 {
			run(fmt.Sprintf("git tag -d %s", tag))
			fmt.Printf("      %s✅ Lokal dihapus%s\n", GR, RS)
		}
	}

	fmt.Println("")
	sep(GR)
	fmt.Printf("%s%s✅  %d tag berhasil dihapus dari %s!%s\n", GR, BD, len(tagsToDelete), scopeLabel, RS)
	sep(GR)
}

// ── Main Menu ─────────────────────────────────────────────────

func showMenu() {
	for {
		clearScreen()
		fmt.Printf("\n%s%s🚀  %s — Release Manager%s\n", CY, BD, projectName, RS)
		sep(DM)
		fmt.Printf("   Branch : %s%s%s   Versi : %sv%s%s\n",
			CY, getCurrentBranch(), RS, YL, getVersion(), RS)
		sep(DM)

		opts := []string{
			CY + "📊  Cek status" + RS,
			GR + "🆕  Release (bump version + commit + push, tag opsional)" + RS,
			YL + "🏷️   Lihat semua tag" + RS,
			DM + "🗑️   Hapus tag" + RS,
			RD + "❌  Keluar" + RS,
		}

		choice := menuSelect("Pilih aksi:", opts)
		if choice == 4 {
			fmt.Printf("\n%s👋 Bye!%s\n\n", GR, RS)
			break
		}

		switch choice {
		case 0:
			run("git fetch --tags --force")
			showStatus()
			showTags()
			pauseBack()
		case 1:
			release()
		case 2:
			run("git fetch --tags --force")
			showTags()
			pauseBack()
		case 3:
			deleteTag()
		}
	}
}

// ── Entry Point ───────────────────────────────────────────────

func main() {
	// Enable ANSI on Windows
	if runtime.GOOS == "windows" {
		exec.Command("cmd", "/c", "").Run()
	}

	arg := "menu"
	if len(os.Args) > 1 {
		arg = os.Args[1]
	}

	switch arg {
	case "status":
		run("git fetch --tags --force")
		showStatus()
		showTags()
	case "release":
		release()
	case "tags":
		run("git fetch --tags --force")
		showTags()
	case "delete", "delete-tag":
		deleteTag()
	default:
		showMenu()
	}
}
