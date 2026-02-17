package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ── Metas ────────────────────────────────────────────────────────────────────

const (
	dailyCommitsTarget   = 20
	dailyIssuesTarget    = 10
	dailyLinesTarget     = 500
	monthlyCommitsTarget = 200
	monthlyIssuesTarget  = 80
	monthlyLinesTarget   = 10000

	tickInterval = 1200 * time.Millisecond
	barWidth     = 28
)

// ── Estilos ──────────────────────────────────────────────────────────────────

var (
	colorPurple = lipgloss.Color("#7D56F4")
	colorFaint  = lipgloss.Color("#3C3C5A")
	colorWhite  = lipgloss.Color("#FAFAFA")
	colorGray   = lipgloss.Color("#6B6B8A")
	colorGreen  = lipgloss.Color("#A8FF78")

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorWhite).
			Background(colorPurple).
			Padding(0, 3)

	sectionTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorPurple).
				BorderStyle(lipgloss.ThickBorder()).
				BorderBottom(true).
				BorderForeground(colorFaint).
				Width(34).
				Align(lipgloss.Center).
				MarginBottom(1)

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#AFAFCF")).
			Bold(true)

	valueStyle = lipgloss.NewStyle().
			Foreground(colorWhite)

	pctStyle = lipgloss.NewStyle().
			Foreground(colorGreen).
			Bold(true)

	colStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Width(38)

	sepStyle = lipgloss.NewStyle().
			Foreground(colorFaint)

	outerStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPurple).
			Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Foreground(colorGray).
			MarginTop(1)

	statusStyle = lipgloss.NewStyle().
			Foreground(colorGray).
			Italic(true)
)

// ── Tipos ─────────────────────────────────────────────────────────────────────

type Metrics struct {
	DailyCommits   int
	MonthlyCommits int
	DailyIssues    int
	MonthlyIssues  int
	DailyLines     int
	MonthlyLines   int
}

type Bars struct {
	DailyCommits   progress.Model
	MonthlyCommits progress.Model
	DailyIssues    progress.Model
	MonthlyIssues  progress.Model
	DailyLines     progress.Model
	MonthlyLines   progress.Model
}

type tickMsg time.Time

type Model struct {
	metrics    Metrics
	bars       Bars
	lastUpdate time.Time
	width      int
	height     int
}

// ── Inicialização ─────────────────────────────────────────────────────────────

func newBar(from, to string) progress.Model {
	return progress.New(
		progress.WithScaledGradient(from, to),
		progress.WithWidth(barWidth),
		progress.WithoutPercentage(),
	)
}

func newModel() Model {
	return Model{
		metrics: Metrics{
			DailyCommits:   8,
			MonthlyCommits: 45,
			DailyIssues:    3,
			MonthlyIssues:  20,
			DailyLines:     200,
			MonthlyLines:   1500,
		},
		bars: Bars{
			DailyCommits:   newBar("#FF6B9D", "#C44569"),
			MonthlyCommits: newBar("#FF6B9D", "#C44569"),
			DailyIssues:    newBar("#4ECDC4", "#1A8A82"),
			MonthlyIssues:  newBar("#4ECDC4", "#1A8A82"),
			DailyLines:     newBar("#45B7D1", "#1A5276"),
			MonthlyLines:   newBar("#45B7D1", "#1A5276"),
		},
		lastUpdate: time.Now(),
	}
}

func pct(current, target int) float64 {
	v := float64(current) / float64(target)
	if v > 1 {
		return 1
	}
	return v
}

// ── Init ──────────────────────────────────────────────────────────────────────

func (m Model) Init() tea.Cmd {
	mx := m.metrics
	return tea.Batch(
		tickCmd(),
		m.bars.DailyCommits.SetPercent(pct(mx.DailyCommits, dailyCommitsTarget)),
		m.bars.MonthlyCommits.SetPercent(pct(mx.MonthlyCommits, monthlyCommitsTarget)),
		m.bars.DailyIssues.SetPercent(pct(mx.DailyIssues, dailyIssuesTarget)),
		m.bars.MonthlyIssues.SetPercent(pct(mx.MonthlyIssues, monthlyIssuesTarget)),
		m.bars.DailyLines.SetPercent(pct(mx.DailyLines, dailyLinesTarget)),
		m.bars.MonthlyLines.SetPercent(pct(mx.MonthlyLines, monthlyLinesTarget)),
	)
}

func tickCmd() tea.Cmd {
	return tea.Tick(tickInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// ── Update ────────────────────────────────────────────────────────────────────

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			m.metrics = Metrics{
				DailyCommits: 1, MonthlyCommits: 5,
				DailyIssues: 1, MonthlyIssues: 2,
				DailyLines: 10, MonthlyLines: 100,
			}
			mx := m.metrics
			return m, tea.Batch(
				m.bars.DailyCommits.SetPercent(pct(mx.DailyCommits, dailyCommitsTarget)),
				m.bars.MonthlyCommits.SetPercent(pct(mx.MonthlyCommits, monthlyCommitsTarget)),
				m.bars.DailyIssues.SetPercent(pct(mx.DailyIssues, dailyIssuesTarget)),
				m.bars.MonthlyIssues.SetPercent(pct(mx.MonthlyIssues, monthlyIssuesTarget)),
				m.bars.DailyLines.SetPercent(pct(mx.DailyLines, dailyLinesTarget)),
				m.bars.MonthlyLines.SetPercent(pct(mx.MonthlyLines, monthlyLinesTarget)),
				tickCmd(),
			)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tickMsg:
		m.lastUpdate = time.Time(msg)
		var cmds []tea.Cmd

		if m.metrics.DailyCommits < dailyCommitsTarget {
			m.metrics.DailyCommits++
			cmds = append(cmds, m.bars.DailyCommits.SetPercent(pct(m.metrics.DailyCommits, dailyCommitsTarget)))
		}
		if m.metrics.MonthlyCommits < monthlyCommitsTarget {
			m.metrics.MonthlyCommits += 3
			cmds = append(cmds, m.bars.MonthlyCommits.SetPercent(pct(m.metrics.MonthlyCommits, monthlyCommitsTarget)))
		}
		if m.metrics.DailyIssues < dailyIssuesTarget {
			m.metrics.DailyIssues++
			cmds = append(cmds, m.bars.DailyIssues.SetPercent(pct(m.metrics.DailyIssues, dailyIssuesTarget)))
		}
		if m.metrics.MonthlyIssues < monthlyIssuesTarget {
			m.metrics.MonthlyIssues += 2
			cmds = append(cmds, m.bars.MonthlyIssues.SetPercent(pct(m.metrics.MonthlyIssues, monthlyIssuesTarget)))
		}
		if m.metrics.DailyLines < dailyLinesTarget {
			m.metrics.DailyLines += 25
			cmds = append(cmds, m.bars.DailyLines.SetPercent(pct(m.metrics.DailyLines, dailyLinesTarget)))
		}
		if m.metrics.MonthlyLines < monthlyLinesTarget {
			m.metrics.MonthlyLines += 200
			cmds = append(cmds, m.bars.MonthlyLines.SetPercent(pct(m.metrics.MonthlyLines, monthlyLinesTarget)))
		}

		cmds = append(cmds, tickCmd())
		return m, tea.Batch(cmds...)

	case progress.FrameMsg:
		var cmds []tea.Cmd

		updateBar := func(b progress.Model) (progress.Model, tea.Cmd) {
			nb, cmd := b.Update(msg)
			return nb.(progress.Model), cmd
		}

		var cmd tea.Cmd
		m.bars.DailyCommits, cmd = updateBar(m.bars.DailyCommits)
		cmds = append(cmds, cmd)
		m.bars.MonthlyCommits, cmd = updateBar(m.bars.MonthlyCommits)
		cmds = append(cmds, cmd)
		m.bars.DailyIssues, cmd = updateBar(m.bars.DailyIssues)
		cmds = append(cmds, cmd)
		m.bars.MonthlyIssues, cmd = updateBar(m.bars.MonthlyIssues)
		cmds = append(cmds, cmd)
		m.bars.DailyLines, cmd = updateBar(m.bars.DailyLines)
		cmds = append(cmds, cmd)
		m.bars.MonthlyLines, cmd = updateBar(m.bars.MonthlyLines)
		cmds = append(cmds, cmd)

		return m, tea.Batch(cmds...)
	}

	return m, nil
}

// ── View ──────────────────────────────────────────────────────────────────────

func metric(icon, label string, bar progress.Model, current, target int) string {
	percentage := pct(current, target) * 100
	row := fmt.Sprintf("%s  %s",
		valueStyle.Render(fmt.Sprintf("%d / %d", current, target)),
		pctStyle.Render(fmt.Sprintf("%.0f%%", percentage)),
	)
	return lipgloss.JoinVertical(lipgloss.Left,
		labelStyle.Render(icon+"  "+label),
		bar.View(),
		row,
	)
}

func sep() string {
	return sepStyle.Render(strings.Repeat("─", barWidth+4))
}

func column(title string, commits, issues, lines metric_args) string {
	return colStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
		sectionTitleStyle.Render(title),
		metric(commits.icon, commits.label, commits.bar, commits.current, commits.target),
		sep(),
		metric(issues.icon, issues.label, issues.bar, issues.current, issues.target),
		sep(),
		metric(lines.icon, lines.label, lines.bar, lines.current, lines.target),
	))
}

type metric_args struct {
	icon, label    string
	bar            progress.Model
	current, target int
}

func (m Model) View() string {
	date := time.Now().Format("02/01/2006")
	title := titleStyle.Render(fmt.Sprintf("  Dev Stats — %s  ", date))

	daily := column("📅  Hoje",
		metric_args{"⎇", "Commits", m.bars.DailyCommits, m.metrics.DailyCommits, dailyCommitsTarget},
		metric_args{"◈", "Issues", m.bars.DailyIssues, m.metrics.DailyIssues, dailyIssuesTarget},
		metric_args{"≡", "Linhas de Código", m.bars.DailyLines, m.metrics.DailyLines, dailyLinesTarget},
	)

	monthly := column("📆  Este Mês",
		metric_args{"⎇", "Commits", m.bars.MonthlyCommits, m.metrics.MonthlyCommits, monthlyCommitsTarget},
		metric_args{"◈", "Issues", m.bars.MonthlyIssues, m.metrics.MonthlyIssues, monthlyIssuesTarget},
		metric_args{"≡", "Linhas de Código", m.bars.MonthlyLines, m.metrics.MonthlyLines, monthlyLinesTarget},
	)

	body := outerStyle.Render(
		lipgloss.JoinHorizontal(lipgloss.Top, daily, monthly),
	)

	status := statusStyle.Render(fmt.Sprintf(
		"  fonte: mock  •  atualizado às %s",
		m.lastUpdate.Format("15:04:05"),
	))
	help := helpStyle.Render("  q sair   r reiniciar")

	return "\n" + lipgloss.JoinVertical(lipgloss.Left,
		title, body, status, help,
	)
}

// ── Main ──────────────────────────────────────────────────────────────────────

func main() {
	p := tea.NewProgram(newModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Erro: %v\n", err)
	}
}
