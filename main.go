package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

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

	tickInterval = 600 * time.Millisecond
	barWidth     = 30
)

// ── GitLab color palette ──────────────────────────────────────────────────────

var (
	glOrange = lipgloss.Color("#FC6D26")
	glPurple = lipgloss.Color("#6B4FBB")
	glWhite  = lipgloss.Color("#FFFFFF")
	glGray   = lipgloss.Color("#8B8FA8")

	// Danger → Warning → Success zones
	zoneRed1    = "#6B0000"
	zoneRed2    = "#DD4132"
	zoneOrange1 = "#C05A00"
	zoneOrange2 = "#FC6D26"
	zoneGreen1  = "#0A5230"
	zoneGreen2  = "#2DA160"

	barEmpty = "#2D2B55"
)

// ── Estilos ──────────────────────────────────────────────────────────────────

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(glWhite).
			Background(glOrange).
			Padding(0, 3)

	sectionTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(glPurple).
				BorderStyle(lipgloss.ThickBorder()).
				BorderBottom(true).
				BorderForeground(glOrange).
				Width(36).
				Align(lipgloss.Center).
				MarginBottom(1)

	labelStyle = lipgloss.NewStyle().
			Foreground(glGray).
			Bold(true)

	valueStyle = lipgloss.NewStyle().
			Foreground(glWhite)

	colStyle = lipgloss.NewStyle().
			Padding(1, 3).
			Width(42)

	sepStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#2D2B55"))

	outerStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(glPurple).
			Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Foreground(glGray).
			MarginTop(1)

	statusStyle = lipgloss.NewStyle().
			Foreground(glGray).
			Italic(true)
)

// ── Banner ANSI Shadow ────────────────────────────────────────────────────────

// "DEV-STATS" em ANSI Shadow — gerado coluna a coluna.
var bannerLines = [6]string{
	`██████╗ ███████╗██╗   ██╗         ███████╗████████╗ █████╗ ████████╗███████╗`,
	`██╔══██╗██╔════╝██║   ██║         ██╔════╝╚══██╔══╝██╔══██╗╚══██╔══╝██╔════╝`,
	`██║  ██║█████╗  ██║   ██║ ─────── ███████╗   ██║   ███████║   ██║   ███████╗`,
	`██║  ██║██╔══╝  ╚██╗ ██╔╝         ╚════██║   ██║   ██╔══██║   ██║   ╚════██║`,
	`██████╔╝███████╗ ╚████╔╝          ███████║   ██║   ██║  ██║   ██║   ███████║`,
	`╚═════╝ ╚══════╝  ╚═══╝           ╚══════╝   ╚═╝   ╚═╝  ╚═╝   ╚═╝   ╚══════╝`,
}

// bannerGrad retorna a cor do gradiente amarelo → laranja → vermelho.
func bannerGrad(t float64) string {
	if t < 0.5 {
		return lerpHex("#FFD700", "#FF6B00", t*2)
	}
	return lerpHex("#FF6B00", "#CC0000", (t-0.5)*2)
}

// renderBanner aplica o gradiente caractere a caractere da esquerda para direita.
func renderBanner(leftPad int) string {
	pad := strings.Repeat(" ", leftPad)
	var sb strings.Builder
	for i, line := range bannerLines {
		sb.WriteString(pad)
		runes := []rune(line)
		n := len(runes)
		for j, ch := range runes {
			t := 0.0
			if n > 1 {
				t = float64(j) / float64(n-1)
			}
			sb.WriteString(
				lipgloss.NewStyle().
					Foreground(lipgloss.Color(bannerGrad(t))).
					Render(string(ch)),
			)
		}
		if i < len(bannerLines)-1 {
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

// ── Gradiente dinâmico ────────────────────────────────────────────────────────

// hexToRGB decompõe uma cor hex em componentes R, G, B (0-255).
func hexToRGB(hex string) (float64, float64, float64) {
	hex = strings.TrimPrefix(hex, "#")
	r, _ := strconv.ParseInt(hex[0:2], 16, 0)
	g, _ := strconv.ParseInt(hex[2:4], 16, 0)
	b, _ := strconv.ParseInt(hex[4:6], 16, 0)
	return float64(r), float64(g), float64(b)
}

// lerpHex interpola linearmente entre duas cores hex.
func lerpHex(from, to string, t float64) string {
	r1, g1, b1 := hexToRGB(from)
	r2, g2, b2 := hexToRGB(to)
	return fmt.Sprintf("#%02X%02X%02X",
		int(r1+t*(r2-r1)),
		int(g1+t*(g2-g1)),
		int(b1+t*(b2-b1)),
	)
}

// zoneColors retorna (startColor, endColor) com base no percentual atual.
// Transição: vermelho (perigo) → laranja GitLab (atenção) → verde (sucesso)
func zoneColors(percent float64) (string, string) {
	switch {
	case percent < 0.4:
		t := percent / 0.4
		return lerpHex(zoneRed1, zoneOrange1, t), lerpHex(zoneRed2, zoneOrange2, t)
	case percent < 0.7:
		t := (percent - 0.4) / 0.3
		return lerpHex(zoneOrange1, zoneGreen1, t), lerpHex(zoneOrange2, zoneGreen2, t)
	default:
		t := (percent - 0.7) / 0.3
		from := lerpHex(zoneGreen1, "#1A7A40", t)
		to := lerpHex(zoneGreen2, "#3EE07F", t)
		return from, to
	}
}

// renderBar desenha a barra com gradiente célula a célula.
func renderBar(current, target int) string {
	p := float64(current) / float64(target)
	if p > 1 {
		p = 1
	}
	filled := int(p * float64(barWidth))
	fromColor, toColor := zoneColors(p)

	var sb strings.Builder
	for i := 0; i < barWidth; i++ {
		if i < filled {
			t := 0.0
			if barWidth > 1 {
				t = float64(i) / float64(barWidth-1)
			}
			color := lerpHex(fromColor, toColor, t)
			sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render("█"))
		} else {
			sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(barEmpty)).Render("░"))
		}
	}
	return sb.String()
}

// pctColor retorna a cor do texto de percentual de acordo com a zona.
func pctColor(p float64) lipgloss.Color {
	switch {
	case p < 0.4:
		return lipgloss.Color("#DD4132")
	case p < 0.7:
		return lipgloss.Color("#FC6D26")
	default:
		return lipgloss.Color("#2DA160")
	}
}

// ── Modelo ────────────────────────────────────────────────────────────────────

type Metrics struct {
	DailyCommits   int
	MonthlyCommits int
	DailyIssues    int
	MonthlyIssues  int
	DailyLines     int
	MonthlyLines   int
}

type tickMsg time.Time

type Model struct {
	metrics    Metrics
	lastUpdate time.Time
	width      int
	height     int
}

func newModel() Model {
	return Model{
		metrics: Metrics{
			DailyCommits:   2,
			MonthlyCommits: 10,
			DailyIssues:    1,
			MonthlyIssues:  5,
			DailyLines:     30,
			MonthlyLines:   300,
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

func (m Model) Init() tea.Cmd {
	return tickCmd()
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
				DailyCommits: 2, MonthlyCommits: 10,
				DailyIssues: 1, MonthlyIssues: 5,
				DailyLines: 30, MonthlyLines: 300,
			}
			return m, tickCmd()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tickMsg:
		m.lastUpdate = time.Time(msg)

		if m.metrics.DailyCommits < dailyCommitsTarget {
			m.metrics.DailyCommits++
		}
		if m.metrics.MonthlyCommits < monthlyCommitsTarget {
			m.metrics.MonthlyCommits += 3
		}
		if m.metrics.DailyIssues < dailyIssuesTarget {
			m.metrics.DailyIssues++
		}
		if m.metrics.MonthlyIssues < monthlyIssuesTarget {
			m.metrics.MonthlyIssues += 2
		}
		if m.metrics.DailyLines < dailyLinesTarget {
			m.metrics.DailyLines += 25
		}
		if m.metrics.MonthlyLines < monthlyLinesTarget {
			m.metrics.MonthlyLines += 200
		}

		return m, tickCmd()
	}

	return m, nil
}

// ── View ──────────────────────────────────────────────────────────────────────

func renderMetric(icon, label string, current, target int) string {
	p := pct(current, target)
	bar := renderBar(current, target)

	pctText := lipgloss.NewStyle().
		Foreground(pctColor(p)).
		Bold(true).
		Render(fmt.Sprintf("%.0f%%", p*100))

	info := fmt.Sprintf("%s  %s",
		valueStyle.Render(fmt.Sprintf("%d / %d", current, target)),
		pctText,
	)

	return lipgloss.JoinVertical(lipgloss.Left,
		labelStyle.Render(icon+"  "+label),
		bar,
		info,
	)
}

func sep() string {
	return sepStyle.Render(strings.Repeat("─", barWidth+6))
}

func renderColumn(title string, commits, issues, lines [2]int) string {
	return colStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
		sectionTitleStyle.Render(title),
		renderMetric("⎇", "Commits", commits[0], commits[1]),
		sep(),
		renderMetric("◈", "Issues", issues[0], issues[1]),
		sep(),
		renderMetric("≡", "Linhas de Código", lines[0], lines[1]),
	))
}

func (m Model) View() string {
	mx := m.metrics
	date := time.Now().Format("02/01/2006")

	// Centraliza o banner horizontalmente
	const bannerWidth = 76 // largura em colunas do ASCII art
	leftPad := 0
	if m.width > bannerWidth {
		leftPad = (m.width - bannerWidth) / 2
	}
	banner := renderBanner(leftPad)

	title := titleStyle.Render(fmt.Sprintf("  Dev Stats — %s  ", date))

	daily := renderColumn("📅  Hoje",
		[2]int{mx.DailyCommits, dailyCommitsTarget},
		[2]int{mx.DailyIssues, dailyIssuesTarget},
		[2]int{mx.DailyLines, dailyLinesTarget},
	)
	monthly := renderColumn("📆  Este Mês",
		[2]int{mx.MonthlyCommits, monthlyCommitsTarget},
		[2]int{mx.MonthlyIssues, monthlyIssuesTarget},
		[2]int{mx.MonthlyLines, monthlyLinesTarget},
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
		banner, title, body, status, help,
	)
}

// ── Main ──────────────────────────────────────────────────────────────────────

func main() {
	p := tea.NewProgram(newModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Erro: %v\n", err)
	}
}
