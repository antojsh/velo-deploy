package tui

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"

	"deploy/internal/config"
	"deploy/internal/deploy"
	"deploy/internal/systemd"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ─── Palette ────────────────────────────────────────────────────────────────

var (
	colorPrimary  = lipgloss.Color("#5FAFFF")
	colorGreen    = lipgloss.Color("#50FA7B")
	colorRed      = lipgloss.Color("#FF5555")
	colorOrange   = lipgloss.Color("#FFB86C")
	colorMuted    = lipgloss.Color("#6272A4")
	colorBgSel    = lipgloss.Color("#313145")
	colorBorder   = lipgloss.Color("#44475A")
	colorWhite    = lipgloss.Color("#F8F8F2")
	colorDim      = lipgloss.Color("#888899")
)

// ─── Styles ─────────────────────────────────────────────────────────────────

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary)

	versionStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1)

	panelFocusStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(0, 1)

	sectionTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorPrimary).
				MarginBottom(0)

	appNameStyle = lipgloss.NewStyle().
			Foreground(colorWhite).
			Bold(true)

	selectedAppStyle = lipgloss.NewStyle().
				Background(colorBgSel).
				Foreground(colorWhite).
				Bold(true)

	activeStyle = lipgloss.NewStyle().
			Foreground(colorGreen).
			Bold(true)

	stoppedStyle = lipgloss.NewStyle().
			Foreground(colorRed)

	labelStyle = lipgloss.NewStyle().
			Foreground(colorDim)

	valueStyle = lipgloss.NewStyle().
			Foreground(colorWhite)

	urlStyle = lipgloss.NewStyle().
			Foreground(colorPrimary)

	helpStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	keyStyle = lipgloss.NewStyle().
			Foreground(colorWhite).
			Background(colorBorder).
			Padding(0, 1)

	errorMsgStyle = lipgloss.NewStyle().
			Foreground(colorOrange)

	successMsgStyle = lipgloss.NewStyle().
			Foreground(colorGreen)

	inputLabelStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	inputActiveStyle = lipgloss.NewStyle().
				Foreground(colorWhite).
				Background(colorBgSel).
				Padding(0, 1)

	inputInactiveStyle = lipgloss.NewStyle().
				Foreground(colorDim).
				Padding(0, 1)

	confirmYesStyle = lipgloss.NewStyle().
			Foreground(colorGreen).
			Bold(true).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorGreen).
			Padding(0, 2)

	confirmNoStyle = lipgloss.NewStyle().
			Foreground(colorRed).
			Bold(true).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorRed).
			Padding(0, 2)

	confirmSelStyle = lipgloss.NewStyle().
			Background(colorBgSel).
			Bold(true).
			Border(lipgloss.RoundedBorder()).
			Padding(0, 2)

	logLineStyle = lipgloss.NewStyle().
			Foreground(colorDim)
)

// ─── States / Panels ────────────────────────────────────────────────────────

const (
	panelLeft  = 0
	panelRight = 1
)

const (
	viewDetail  = "detail"
	viewLogs    = "logs"
	viewNew     = "new"
	viewAdd     = "add"
	viewConfirm = "confirm"
	viewOutput  = "output"
)

// ─── Input field helpers ─────────────────────────────────────────────────────

type inputField struct {
	label       string
	value       string
	placeholder string
	focused     bool
}

func (f inputField) render(width int) string {
	label := inputLabelStyle.Render(f.label)
	display := f.value
	if display == "" {
		display = f.placeholder
	}
	cursor := ""
	if f.focused {
		cursor = "█"
	}
	var field string
	if f.focused {
		field = inputActiveStyle.Width(width - 4).Render(display + cursor)
	} else {
		field = inputInactiveStyle.Width(width - 4).Render(display)
	}
	return label + "\n" + field
}

// ─── Messages ────────────────────────────────────────────────────────────────

type tickMsg time.Time
type logsLoadedMsg string
type statusLoadedMsg string
type deployOutputMsg string
type deployDoneMsg struct{ err error }

// ─── Model ───────────────────────────────────────────────────────────────────

type model struct {
	cfg     *config.Config
	apps    []string // sorted app names
	cursor  int
	focus   int    // panelLeft or panelRight
	view    string // active right-panel view

	// terminal size
	width  int
	height int

	// logs
	logsLines  []string
	logsScroll int

	// systemd status text
	statusText string

	// new-deploy form
	formFields  [3]inputField
	formFocus   int // 0=repo 1=domain 2=alias

	// add-existing form
	addFields   [4]inputField
	addFocus    int // 0=name 1=dir 2=domain 3=alias/type

	// confirm delete
	confirmSel int // 0=yes 1=no

	// deploy output streaming
	deployOutput []string
	deploying    bool

	// notification bar
	notification     string
	notificationKind string // "ok" | "err"
}

func NewProgram(cfg *config.Config) *tea.Program {
	return tea.NewProgram(initialModel(cfg), tea.WithAltScreen())
}

func initialModel(cfg *config.Config) model {
	names := sortedApps(cfg)
	m := model{
		cfg:    cfg,
		apps:   names,
		focus:  panelLeft,
		view:   viewDetail,
		width:  120,
		height: 40,
	}
	m.formFields = [3]inputField{
		{label: "Repo URL", placeholder: "https://github.com/user/repo"},
		{label: "Domain  (optional)", placeholder: "api.example.com"},
		{label: "Alias   (optional)", placeholder: "myapp.local"},
	}
	return m
}

func sortedApps(cfg *config.Config) []string {
	names := make([]string, 0, len(cfg.Apps))
	for n := range cfg.Apps {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// ─── Init ────────────────────────────────────────────────────────────────────

func (m model) Init() tea.Cmd {
	return tick()
}

func tick() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// ─── Update ──────────────────────────────────────────────────────────────────

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tickMsg:
		return m, tick()

	case logsLoadedMsg:
		m.logsLines = strings.Split(string(msg), "\n")
		// scroll to bottom
		visibleLines := m.rightPanelHeight() - 4
		if len(m.logsLines) > visibleLines {
			m.logsScroll = len(m.logsLines) - visibleLines
		}
		return m, nil

	case statusLoadedMsg:
		m.statusText = string(msg)
		return m, nil

	case deployOutputMsg:
		m.deployOutput = append(m.deployOutput, string(msg))
		return m, nil

	case deployDoneMsg:
		m.deploying = false
		if msg.err != nil {
			m = m.withNotification("Deploy failed: "+msg.err.Error(), "err")
		} else {
			m = m.withNotification("Deploy successful!", "ok")
			// Reload config & app list
			m.apps = sortedApps(m.cfg)
		}
		m.view = viewOutput
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m model) withNotification(msg, kind string) model {
	m.notification = msg
	m.notificationKind = kind
	return m
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Global quit
	if key == "ctrl+c" {
		return m, tea.Quit
	}

	// ── New deploy form ──────────────────────────────────────────────────────
	if m.view == viewNew {
		return m.handleFormKey(key)
	}

	// ── Add existing app form ─────────────────────────────────────────────────
	if m.view == viewAdd {
		return m.handleAddKey(key)
	}

	// ── Confirm delete ───────────────────────────────────────────────────────
	if m.view == viewConfirm {
		return m.handleConfirmKey(key)
	}

	// ── Logs scroll (right panel focused) ───────────────────────────────────
	if m.view == viewLogs && m.focus == panelRight {
		switch key {
		case "up", "k":
			if m.logsScroll > 0 {
				m.logsScroll--
			}
			return m, nil
		case "down", "j":
			maxScroll := len(m.logsLines) - (m.rightPanelHeight() - 4)
			if maxScroll < 0 {
				maxScroll = 0
			}
			if m.logsScroll < maxScroll {
				m.logsScroll++
			}
			return m, nil
		case "esc", "q":
			m.view = viewDetail
			m.focus = panelLeft
			return m, nil
		}
	}

	// ── Output view ──────────────────────────────────────────────────────────
	if m.view == viewOutput {
		if key == "esc" || key == "q" || key == "enter" {
			m.view = viewDetail
			m.focus = panelLeft
			m.deployOutput = nil
		}
		return m, nil
	}

	// ── Tab: toggle panel focus ───────────────────────────────────────────────
	if key == "tab" {
		if m.focus == panelLeft {
			m.focus = panelRight
		} else {
			m.focus = panelLeft
		}
		return m, nil
	}

	// ── Left panel navigation ─────────────────────────────────────────────────
	if m.focus == panelLeft || m.view == viewDetail {
		switch key {
		case "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.view = viewDetail
				m.statusText = ""
				app := m.cfg.GetApp(m.currentApp())
				if app != nil && app.Type != config.AppTypeStatic {
					return m, loadStatus(m.currentApp())
				}
				return m, nil
			}
		case "down", "j":
			if m.cursor < len(m.apps)-1 {
				m.cursor++
				m.view = viewDetail
				m.statusText = ""
				app := m.cfg.GetApp(m.currentApp())
				if app != nil && app.Type != config.AppTypeStatic {
					return m, loadStatus(m.currentApp())
				}
				return m, nil
			}
		case "enter":
			if len(m.apps) > 0 {
				m.focus = panelRight
			}
		case "esc":
			m.view = viewDetail
			m.focus = panelLeft
		case "l", "L":
			if len(m.apps) > 0 {
				m.view = viewLogs
				m.logsScroll = 0
				m.logsLines = nil
				m.focus = panelRight
				return m, loadLogs(m.currentApp())
			}
		case "r", "R":
			if len(m.apps) > 0 {
				go systemd.RestartApp(m.currentApp())
				m = m.withNotification("Restarting "+m.currentApp()+"…", "ok")
			}
		case "s", "S":
			if len(m.apps) > 0 {
				go systemd.StopApp(m.currentApp())
				m = m.withNotification("Stopping "+m.currentApp()+"…", "ok")
			}
		case "d", "D":
			if len(m.apps) > 0 {
				m.view = viewConfirm
				m.confirmSel = 1 // default: No
				m.focus = panelRight
			}
		case "n", "N":
			m.view = viewNew
			m.focus = panelRight
			m.formFocus = 0
			m.formFields[0].value = ""
			m.formFields[1].value = ""
			m.formFields[2].value = ""
			for i := range m.formFields {
				m.formFields[i].focused = i == 0
			}
		case "a", "A":
			m.view = viewAdd
			m.focus = panelRight
			m.addFocus = 0
			m.addFields = [4]inputField{
				{label: "Name", placeholder: "my-static-site"},
				{label: "Directory", placeholder: "/opt/deploy/apps/my-site"},
				{label: "Domain  (optional)", placeholder: "site.example.com"},
				{label: "Type    (optional)", placeholder: "auto (static/node)"},
			}
			for i := range m.addFields {
				m.addFields[i].focused = i == 0
			}
		}
	}

	return m, nil
}

func (m model) handleFormKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc":
		m.view = viewDetail
		m.focus = panelLeft
	case "tab", "down":
		m.formFields[m.formFocus].focused = false
		m.formFocus = (m.formFocus + 1) % len(m.formFields)
		m.formFields[m.formFocus].focused = true
	case "shift+tab", "up":
		m.formFields[m.formFocus].focused = false
		m.formFocus = (m.formFocus + len(m.formFields) - 1) % len(m.formFields)
		m.formFields[m.formFocus].focused = true
	case "enter":
		if m.formFocus < len(m.formFields)-1 {
			// advance to next field
			m.formFields[m.formFocus].focused = false
			m.formFocus++
			m.formFields[m.formFocus].focused = true
		} else {
			// submit
			return m.submitDeploy()
		}
	case "backspace":
		f := &m.formFields[m.formFocus]
		if len(f.value) > 0 {
			f.value = f.value[:len(f.value)-1]
		}
	default:
		if len(key) == 1 {
			m.formFields[m.formFocus].value += key
		}
	}
	return m, nil
}

func (m model) submitDeploy() (tea.Model, tea.Cmd) {
	repoURL := strings.TrimSpace(m.formFields[0].value)
	if repoURL == "" {
		m = m.withNotification("Repo URL is required", "err")
		return m, nil
	}
	domain := strings.TrimSpace(m.formFields[1].value)
	alias := strings.TrimSpace(m.formFields[2].value)

	m.deploying = true
	m.deployOutput = []string{"Starting deploy for " + repoURL + "…"}
	m.view = viewOutput
	m.focus = panelRight

	cfg := m.cfg
	return m, func() tea.Msg {
		err := deploy.Deploy(cfg, repoURL, domain, alias)
		return deployDoneMsg{err: err}
	}
}

func (m model) handleAddKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc":
		m.view = viewDetail
		m.focus = panelLeft
	case "tab", "down":
		m.addFields[m.addFocus].focused = false
		m.addFocus = (m.addFocus + 1) % len(m.addFields)
		m.addFields[m.addFocus].focused = true
	case "shift+tab", "up":
		m.addFields[m.addFocus].focused = false
		m.addFocus = (m.addFocus + len(m.addFields) - 1) % len(m.addFields)
		m.addFields[m.addFocus].focused = true
	case "enter":
		if m.addFocus < len(m.addFields)-1 {
			m.addFields[m.addFocus].focused = false
			m.addFocus++
			m.addFields[m.addFocus].focused = true
		} else {
			return m.submitAdd()
		}
	case "backspace":
		f := &m.addFields[m.addFocus]
		if len(f.value) > 0 {
			f.value = f.value[:len(f.value)-1]
		}
	default:
		if len(key) == 1 {
			m.addFields[m.addFocus].value += key
		}
	}
	return m, nil
}

func (m model) submitAdd() (tea.Model, tea.Cmd) {
	appName := strings.TrimSpace(m.addFields[0].value)
	appDir := strings.TrimSpace(m.addFields[1].value)
	domain := strings.TrimSpace(m.addFields[2].value)
	typeHint := strings.TrimSpace(m.addFields[3].value)

	if appName == "" || appDir == "" {
		m = m.withNotification("Name and Directory are required", "err")
		return m, nil
	}

	if appDir == "<app-dir>" || strings.HasPrefix(appDir, "/opt/deploy/apps") == false {
		// Use default path if not specified
		if appDir == "<app-dir>" || appDir == "" {
			appDir = "/opt/deploy/apps/" + appName
		}
	}

	appType := ""
	if typeHint != "" && typeHint != "auto (static/node)" {
		if typeHint == "static" {
			appType = config.AppTypeStatic
		} else if typeHint == "node" {
			appType = config.AppTypeNode
		}
	}

	m.deploying = true
	m.deployOutput = []string{"Registering " + appName + " from " + appDir + "…"}
	m.view = viewOutput
	m.focus = panelRight

	cfg := m.cfg
	return m, func() tea.Msg {
		err := deploy.Register(cfg, appName, appDir, domain, "", appType)
		return deployDoneMsg{err: err}
	}
}

func (m model) handleConfirmKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "left", "h", "right", "l":
		m.confirmSel = 1 - m.confirmSel
	case "y", "Y":
		m.confirmSel = 0
		return m.doDelete()
	case "n", "N", "esc":
		m.view = viewDetail
		m.focus = panelLeft
	case "enter":
		if m.confirmSel == 0 {
			return m.doDelete()
		}
		m.view = viewDetail
		m.focus = panelLeft
	}
	return m, nil
}

func (m model) doDelete() (tea.Model, tea.Cmd) {
	name := m.currentApp()
	m.view = viewOutput
	m.focus = panelRight
	m.deployOutput = []string{"Removing " + name + "…"}
	cfg := m.cfg

	return m, func() tea.Msg {
		err := deploy.Remove(cfg, name)
		return deployDoneMsg{err: err}
	}
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func (m model) currentApp() string {
	if len(m.apps) == 0 {
		return ""
	}
	return m.apps[m.cursor]
}

func (m model) leftPanelWidth() int {
	w := m.width / 4
	if w < 22 {
		w = 22
	}
	if w > 36 {
		w = 36
	}
	return w
}

func (m model) rightPanelWidth() int {
	return m.width - m.leftPanelWidth() - 3
}

func (m model) rightPanelHeight() int {
	return m.height - 5 // title + help bars
}

// ─── Async commands ──────────────────────────────────────────────────────────

func loadLogs(appName string) tea.Cmd {
	return func() tea.Msg {
		out, _ := exec.Command("journalctl", "-u", "deploy-"+appName+".service",
			"--no-pager", "-n", "200", "--output=short").CombinedOutput()
		return logsLoadedMsg(out)
	}
}

func loadStatus(appName string) tea.Cmd {
	if appName == "" {
		return nil
	}
	return func() tea.Msg {
		out, _ := exec.Command("systemctl", "status", "deploy-"+appName+".service",
			"--no-pager", "-l").CombinedOutput()
		return statusLoadedMsg(out)
	}
}

// ─── View ────────────────────────────────────────────────────────────────────

func (m model) View() string {
	var out strings.Builder

	// ── Title bar ─────────────────────────────────────────────────────────────
	title := titleStyle.Render("⚡ deploy")
	ver := versionStyle.Render(" — Bare Metal PaaS")
	notif := ""
	if m.notification != "" {
		if m.notificationKind == "err" {
			notif = "  " + errorMsgStyle.Render("✗ "+m.notification)
		} else {
			notif = "  " + successMsgStyle.Render("✓ "+m.notification)
		}
	}
	titleBar := title + ver + notif
	out.WriteString(titleBar + "\n\n")

	// ── Panels ────────────────────────────────────────────────────────────────
	left := m.renderLeft()
	right := m.renderRight()

	h := m.rightPanelHeight()
	lw := m.leftPanelWidth()
	rw := m.rightPanelWidth()

	leftPanel := m.applyPanel(left, lw, h, m.focus == panelLeft)
	rightPanel := m.applyPanel(right, rw, h, m.focus == panelRight)

	panels := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, "  ", rightPanel)
	out.WriteString(panels + "\n")

	// ── Help bar ──────────────────────────────────────────────────────────────
	out.WriteString("\n" + m.renderHelp())

	return out.String()
}

func (m model) applyPanel(content string, w, h int, focused bool) string {
	style := panelStyle.Width(w).Height(h)
	if focused {
		style = panelFocusStyle.Width(w).Height(h)
	}
	return style.Render(content)
}

// ─── Left panel: app list ────────────────────────────────────────────────────

func (m model) renderLeft() string {
	var s strings.Builder
	s.WriteString(sectionTitleStyle.Render("  APPS") + "\n\n")

	if len(m.apps) == 0 {
		s.WriteString(labelStyle.Render("  No apps deployed.\n  Press [N] to deploy."))
		return s.String()
	}

	for i, name := range m.apps {
		active := systemd.IsActive(name)
		dot := stoppedStyle.Render("○")
		if active {
			dot = activeStyle.Render("●")
		}

		app := m.cfg.Apps[name]
		icon := "🔧"
		if app != nil && app.Type == config.AppTypeStatic {
			icon = "🌐"
		}

		line := fmt.Sprintf(" %s %s %s", dot, icon, name)
		if i == m.cursor {
			s.WriteString(selectedAppStyle.Render(fmt.Sprintf("%-*s", m.leftPanelWidth()-2, line)) + "\n")
		} else {
			s.WriteString(appNameStyle.Render(line) + "\n")
		}
	}

	return s.String()
}

// ─── Right panel dispatcher ──────────────────────────────────────────────────

func (m model) renderRight() string {
	switch m.view {
	case viewLogs:
		return m.renderLogs()
	case viewNew:
		return m.renderNewForm()
	case viewAdd:
		return m.renderAddForm()
	case viewConfirm:
		return m.renderConfirm()
	case viewOutput:
		return m.renderOutput()
	default:
		return m.renderDetail()
	}
}

// ─── Detail view ─────────────────────────────────────────────────────────────

func (m model) renderDetail() string {
	if len(m.apps) == 0 {
		return labelStyle.Render("No app selected.\nPress [N] to deploy your first app.")
	}

	name := m.currentApp()
	app := m.cfg.Apps[name]
	if app == nil {
		return labelStyle.Render("App not found.")
	}

	active := systemd.IsActive(name)
	statusStr := stoppedStyle.Render("○ stopped")
	if active {
		statusStr = activeStyle.Render("● active")
	}

	url := fmt.Sprintf("https://<server-ip>/%s/", name)
	if app.Domain != "" {
		url = "https://" + app.Domain
	}

	isStatic := app.Type == config.AppTypeStatic

	var s strings.Builder

	s.WriteString(sectionTitleStyle.Render("APP INFO") + "\n")
	s.WriteString(strings.Repeat("─", m.rightPanelWidth()-4) + "\n")
	s.WriteString(row("Name    ", app.Name) + "\n")
	s.WriteString(row("Type    ", app.Type) + "\n")
	s.WriteString(rowRaw("Status  ", statusStr) + "\n")
	s.WriteString(rowURL("URL     ", url) + "\n")
	if app.RepoURL != "" {
		s.WriteString(row("Repo    ", app.RepoURL) + "\n")
	}
	if app.Branch != "" {
		s.WriteString(row("Branch  ", app.Branch) + "\n")
	}
	if app.NodeVer != "" {
		s.WriteString(row("Node    ", app.NodeVer) + "\n")
	}
	if app.Port != 0 {
		s.WriteString(rowInt("Port    ", app.Port) + "\n")
	}
	if app.Alias != "" {
		s.WriteString(row("Alias   ", app.Alias) + "\n")
	}
	if app.EntryPoint != "" {
		s.WriteString(row("Entry   ", app.EntryPoint) + "\n")
	}
	if app.OutputDir != "" {
		s.WriteString(row("Output  ", app.OutputDir) + "\n")
	}
	s.WriteString("\n")

	if isStatic {
		s.WriteString(sectionTitleStyle.Render("SERVICE") + "\n")
		s.WriteString(strings.Repeat("─", m.rightPanelWidth()-4) + "\n")
		s.WriteString(labelStyle.Render("  Caddy (static file_server)") + "\n")
		s.WriteString("\n")
	} else {
		s.WriteString(sectionTitleStyle.Render("SYSTEMD") + "\n")
		s.WriteString(strings.Repeat("─", m.rightPanelWidth()-4) + "\n")
		if m.statusText == "" {
			s.WriteString(labelStyle.Render("  Loading…\n"))
		} else {
			lines := strings.Split(m.statusText, "\n")
			maxLines := 8
			if len(lines) < maxLines {
				maxLines = len(lines)
			}
			for _, l := range lines[:maxLines] {
				s.WriteString(logLineStyle.Render("  "+truncate(l, m.rightPanelWidth()-6)) + "\n")
			}
		}
		s.WriteString("\n")
	}

	s.WriteString(sectionTitleStyle.Render("ACTIONS") + "\n")
	s.WriteString(strings.Repeat("─", m.rightPanelWidth()-4) + "\n")
	if isStatic {
		s.WriteString(
			keyStyle.Render("L") + helpStyle.Render(" Logs  ") +
				keyStyle.Render("D") + helpStyle.Render(" Delete") + "\n",
		)
	} else {
		s.WriteString(
			keyStyle.Render("R") + helpStyle.Render(" Restart  ") +
				keyStyle.Render("S") + helpStyle.Render(" Stop  ") +
				keyStyle.Render("L") + helpStyle.Render(" Logs  ") +
				keyStyle.Render("D") + helpStyle.Render(" Delete") + "\n",
		)
	}

	return s.String()
}

// ─── Logs view ───────────────────────────────────────────────────────────────

func (m model) renderLogs() string {
	var s strings.Builder

	name := m.currentApp()
	s.WriteString(sectionTitleStyle.Render("LOGS — "+name) + "\n")
	s.WriteString(strings.Repeat("─", m.rightPanelWidth()-4) + "\n\n")

	if len(m.logsLines) == 0 {
		s.WriteString(labelStyle.Render("Loading…"))
		return s.String()
	}

	visible := m.rightPanelHeight() - 6
	start := m.logsScroll
	end := start + visible
	if end > len(m.logsLines) {
		end = len(m.logsLines)
	}

	for _, l := range m.logsLines[start:end] {
		s.WriteString(logLineStyle.Render(truncate(l, m.rightPanelWidth()-4)) + "\n")
	}

	// scroll indicator
	total := len(m.logsLines)
	s.WriteString("\n" + labelStyle.Render(fmt.Sprintf("Lines %d-%d of %d  [↑/↓ scroll]", start+1, end, total)))

	return s.String()
}

// ─── New deploy form ─────────────────────────────────────────────────────────

func (m model) renderNewForm() string {
	var s strings.Builder

	s.WriteString(sectionTitleStyle.Render("NEW DEPLOY") + "\n")
	s.WriteString(strings.Repeat("─", m.rightPanelWidth()-4) + "\n\n")

	fw := m.rightPanelWidth() - 6
	for _, f := range m.formFields {
		s.WriteString(f.render(fw) + "\n\n")
	}

	s.WriteString("\n")
	deployBtn := keyStyle.Render("Enter") + helpStyle.Render(" Deploy   ") +
		keyStyle.Render("Tab") + helpStyle.Render(" Next field   ") +
		keyStyle.Render("Esc") + helpStyle.Render(" Cancel")
	s.WriteString(deployBtn)

	if m.notification != "" && m.notificationKind == "err" {
		s.WriteString("\n\n" + errorMsgStyle.Render("✗ "+m.notification))
	}

	return s.String()
}

// ─── Add existing app form ────────────────────────────────────────────────────

func (m model) renderAddForm() string {
	var s strings.Builder

	s.WriteString(sectionTitleStyle.Render("ADD EXISTING APP") + "\n")
	s.WriteString(strings.Repeat("─", m.rightPanelWidth()-4) + "\n\n")

	fw := m.rightPanelWidth() - 6
	for _, f := range m.addFields {
		s.WriteString(f.render(fw) + "\n\n")
	}

	s.WriteString("\n")
	addBtn := keyStyle.Render("Enter") + helpStyle.Render(" Add     ") +
		keyStyle.Render("Tab") + helpStyle.Render(" Next field   ") +
		keyStyle.Render("Esc") + helpStyle.Render(" Cancel")
	s.WriteString(addBtn)

	if m.notification != "" && m.notificationKind == "err" {
		s.WriteString("\n\n" + errorMsgStyle.Render("✗ "+m.notification))
	}

	return s.String()
}

// ─── Confirm delete ──────────────────────────────────────────────────────────

func (m model) renderConfirm() string {
	name := m.currentApp()
	var s strings.Builder

	s.WriteString(sectionTitleStyle.Render("CONFIRM DELETE") + "\n")
	s.WriteString(strings.Repeat("─", m.rightPanelWidth()-4) + "\n\n")

	s.WriteString(errorMsgStyle.Render(fmt.Sprintf(
		"  Are you sure you want to delete '%s'?\n  This will stop the service, remove the files,\n  and delete the system user.\n\n",
		name,
	)))

	yes := "  YES  "
	no := "  NO  "
	if m.confirmSel == 0 {
		s.WriteString(confirmSelStyle.Foreground(colorRed).Render(yes))
	} else {
		s.WriteString(confirmYesStyle.Render(yes))
	}
	s.WriteString("   ")
	if m.confirmSel == 1 {
		s.WriteString(confirmSelStyle.Foreground(colorGreen).Render(no))
	} else {
		s.WriteString(confirmNoStyle.Render(no))
	}

	s.WriteString("\n\n" + helpStyle.Render("[←/→] Select  [Enter] Confirm  [Esc] Cancel"))
	return s.String()
}

// ─── Output / deploy log view ────────────────────────────────────────────────

func (m model) renderOutput() string {
	var s strings.Builder

	title := "DEPLOY OUTPUT"
	if !m.deploying && len(m.deployOutput) > 0 {
		title = "OPERATION COMPLETE"
	}
	s.WriteString(sectionTitleStyle.Render(title) + "\n")
	s.WriteString(strings.Repeat("─", m.rightPanelWidth()-4) + "\n\n")

	visible := m.rightPanelHeight() - 7
	lines := m.deployOutput
	start := 0
	if len(lines) > visible {
		start = len(lines) - visible
	}
	for _, l := range lines[start:] {
		s.WriteString(logLineStyle.Render(truncate(l, m.rightPanelWidth()-4)) + "\n")
	}

	if m.deploying {
		s.WriteString("\n" + labelStyle.Render("Running…"))
	} else {
		s.WriteString("\n" + helpStyle.Render("[Esc / Enter] Back"))
	}

	return s.String()
}

// ─── Help bar ────────────────────────────────────────────────────────────────

func (m model) renderHelp() string {
	base := helpStyle.Render(
		"[↑/↓] Navigate  [Tab] Switch panel  [N] New deploy  [A] Add existing  [Q] Quit",
	)
	return base
}

// ─── Formatting helpers ───────────────────────────────────────────────────────

func row(label, value string) string {
	return labelStyle.Render("  "+label+": ") + valueStyle.Render(value)
}

func rowRaw(label, value string) string {
	return labelStyle.Render("  "+label+": ") + value
}

func rowURL(label, value string) string {
	return labelStyle.Render("  "+label+": ") + urlStyle.Render(value)
}

func rowInt(label string, value int) string {
	return labelStyle.Render("  "+label+": ") + valueStyle.Render(fmt.Sprintf("%d", value))
}

func truncate(s string, max int) string {
	if max <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) > max {
		return string(runes[:max-1]) + "…"
	}
	return s
}
