package tui

import (
    "errors"
    "testing"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/stretchr/testify/assert"

    "velo-deploy/internal/config"
)

func testConfig() *config.Config {
    return &config.Config{
        AppsDir: "/tmp/apps",
        Apps: map[string]*config.AppMeta{
            "zeta": {Name: "zeta", Type: config.AppTypeNode, Port: 3001, Alias: "zeta.local", NodeVer: "24", EntryPoint: "server.js"},
            "alpha": {Name: "alpha", Type: config.AppTypeStatic, Domain: "alpha.example.com", OutputDir: "dist"},
        },
    }
}

func asModel(t *testing.T, tm tea.Model) model {
    t.Helper()
    m, ok := tm.(model)
    if !ok {
        t.Fatalf("expected tui model, got %T", tm)
    }
    return m
}

func TestInitialModelSortsAppsAndDefaults(t *testing.T) {
    m := initialModel(testConfig())

    assert.Equal(t, []string{"alpha", "zeta"}, m.apps)
    assert.Equal(t, panelLeft, m.focus)
    assert.Equal(t, viewDetail, m.view)
    assert.Equal(t, 120, m.width)
    assert.Equal(t, 40, m.height)
    assert.Equal(t, "Repo URL", m.formFields[0].label)
}

func TestUpdateWindowSizeAndMessages(t *testing.T) {
    m := initialModel(testConfig())

    updated := asModel(t, mustModel(m.Update(tea.WindowSizeMsg{Width: 90, Height: 30})))
    assert.Equal(t, 90, updated.width)
    assert.Equal(t, 30, updated.height)

    updated = asModel(t, mustModel(updated.Update(statusLoadedMsg("active\nloaded"))))
    assert.Equal(t, "active\nloaded", updated.statusText)

    updated = asModel(t, mustModel(updated.Update(deployOutputMsg("line one"))))
    assert.Equal(t, []string{"line one"}, updated.deployOutput)

    updated = asModel(t, mustModel(updated.Update(logsLoadedMsg("a\nb\nc"))))
    assert.Equal(t, []string{"a", "b", "c"}, updated.logsLines)
}

func TestDeployDoneSuccessReloadsAppsAndShowsOutput(t *testing.T) {
    cfg := &config.Config{Apps: map[string]*config.AppMeta{}}
    m := initialModel(cfg)
    m.deploying = true
    cfg.Apps["newapp"] = &config.AppMeta{Name: "newapp"}

    updated := asModel(t, mustModel(m.Update(deployDoneMsg{})))

    assert.False(t, updated.deploying)
    assert.Equal(t, viewOutput, updated.view)
    assert.Equal(t, "ok", updated.notificationKind)
    assert.Equal(t, []string{"newapp"}, updated.apps)
}

func TestDeployDoneErrorKeepsErrorNotification(t *testing.T) {
    m := initialModel(testConfig())
    m.deploying = true

    updated := asModel(t, mustModel(m.Update(deployDoneMsg{err: errors.New("boom")})))

    assert.False(t, updated.deploying)
    assert.Equal(t, viewOutput, updated.view)
    assert.Equal(t, "err", updated.notificationKind)
    assert.Contains(t, updated.notification, "boom")
}

func TestHandleKeyNavigationAndViews(t *testing.T) {
    m := initialModel(testConfig())

    updated := asModel(t, mustModel(m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})))
    assert.Equal(t, 1, updated.cursor)

    updated = asModel(t, mustModel(updated.handleKey(tea.KeyMsg{Type: tea.KeyTab})))
    assert.Equal(t, panelRight, updated.focus)

    updated = asModel(t, mustModel(updated.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})))
    assert.Equal(t, viewNew, updated.view)
    assert.True(t, updated.formFields[0].focused)

    updated = asModel(t, mustModel(updated.handleKey(tea.KeyMsg{Type: tea.KeyEsc})))
    assert.Equal(t, viewDetail, updated.view)
    assert.Equal(t, panelLeft, updated.focus)
}

func TestFormKeyEditingAndValidation(t *testing.T) {
    m := initialModel(testConfig())
    m.view = viewNew
    m.formFocus = 0
    m.formFields[0].focused = true

    updated := asModel(t, mustModel(m.handleFormKey("x")))
    assert.Equal(t, "x", updated.formFields[0].value)

    updated = asModel(t, mustModel(updated.handleFormKey("backspace")))
    assert.Equal(t, "", updated.formFields[0].value)

    updated = asModel(t, mustModel(updated.submitDeploy()))
    assert.Equal(t, "err", updated.notificationKind)
    assert.Contains(t, updated.notification, "Repo URL is required")
}

func TestAddFormValidationAndFieldMovement(t *testing.T) {
    m := initialModel(testConfig())
    m.view = viewAdd
    m.addFields[0].focused = true

    updated := asModel(t, mustModel(m.handleAddKey("a")))
    assert.Equal(t, "a", updated.addFields[0].value)

    updated = asModel(t, mustModel(updated.handleAddKey("tab")))
    assert.Equal(t, 1, updated.addFocus)
    assert.True(t, updated.addFields[1].focused)

    updated.addFocus = len(updated.addFields) - 1
    updated.addFields[0].value = ""
    updated = asModel(t, mustModel(updated.submitAdd()))
    assert.Equal(t, "err", updated.notificationKind)
    assert.Contains(t, updated.notification, "Name and Directory are required")
}

func TestConfirmKeyCancelAndToggle(t *testing.T) {
    m := initialModel(testConfig())
    m.view = viewConfirm
    m.focus = panelRight
    m.confirmSel = 1

    updated := asModel(t, mustModel(m.handleConfirmKey("left")))
    assert.Equal(t, 0, updated.confirmSel)

    updated = asModel(t, mustModel(updated.handleConfirmKey("n")))
    assert.Equal(t, viewDetail, updated.view)
    assert.Equal(t, panelLeft, updated.focus)
}

func TestCurrentAppAndDimensions(t *testing.T) {
    m := initialModel(testConfig())
    assert.Equal(t, "alpha", m.currentApp())

    empty := initialModel(&config.Config{Apps: map[string]*config.AppMeta{}})
    assert.Equal(t, "", empty.currentApp())

    m.width = 40
    assert.Equal(t, 22, m.leftPanelWidth())
    m.width = 200
    assert.Equal(t, 36, m.leftPanelWidth())
    assert.Equal(t, m.width-m.leftPanelWidth()-3, m.rightPanelWidth())

    m.height = 50
    assert.Equal(t, 45, m.rightPanelHeight())
}

func TestRenderingHelpers(t *testing.T) {
    assert.Equal(t, "", truncate("hello", 0))
    assert.Equal(t, "hello", truncate("hello", 10))
    assert.NotEqual(t, "hello", truncate("hello", 4))

    assert.Contains(t, row("Name", "app"), "app")
    assert.Contains(t, rowRaw("Status", "active"), "active")
    assert.Contains(t, rowURL("URL", "https://example.com"), "https://example.com")
    assert.Contains(t, rowInt("Port", 3000), "3000")

    field := inputField{label: "Repo", placeholder: "url", focused: true}
    assert.Contains(t, field.render(20), "Repo")
    assert.Contains(t, field.render(20), "url")
}

func TestRenderViewsDoNotPanic(t *testing.T) {
    m := initialModel(&config.Config{Apps: map[string]*config.AppMeta{}})
    assert.Contains(t, m.renderDetail(), "No app selected")
    assert.Contains(t, m.renderLeft(), "No apps deployed")
    assert.Contains(t, m.renderHelp(), "Navigate")

    m.view = viewOutput
    m.deployOutput = []string{"one", "two"}
    assert.Contains(t, m.renderOutput(), "one")

    m.view = viewLogs
    m.logsLines = []string{"log1", "log2"}
    assert.Contains(t, m.renderLogs(), "log1")

    m.view = viewNew
    assert.Contains(t, m.renderNewForm(), "NEW DEPLOY")

    m.view = viewAdd
    m.addFields = [4]inputField{{label: "Name"}, {label: "Directory"}, {label: "Domain"}, {label: "Type"}}
    assert.Contains(t, m.renderAddForm(), "ADD EXISTING APP")

    m.view = viewConfirm
    assert.Contains(t, m.renderConfirm(), "CONFIRM DELETE")
}

func TestLoadStatusEmptyAppReturnsNilCommand(t *testing.T) {
    assert.Nil(t, loadStatus(""))
}

func mustModel(tm tea.Model, _ tea.Cmd) tea.Model {
    return tm
}
