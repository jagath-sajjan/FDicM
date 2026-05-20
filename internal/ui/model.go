package ui

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"fdicm/internal/api"
	"fdicm/internal/audio"
	"fdicm/internal/cache"
	"fdicm/internal/spell"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type state int

const (
	stateInput state = iota
	stateResult
	stateHistory
)

var (
	purple = lipgloss.Color("#9876ff")
	cyan   = lipgloss.Color("#00f5d4")
	gray   = lipgloss.Color("#444444")

	stylePane = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(gray).
			Padding(1, 2)

	styleActivePane = stylePane.Copy().
			BorderForeground(purple)

	styleTitle = lipgloss.NewStyle().
			Background(purple).
			Foreground(lipgloss.Color("#000000")).
			Bold(true).
			Padding(0, 1)

	styleAccent = lipgloss.NewStyle().Foreground(cyan).Bold(true)
	styleDim    = lipgloss.NewStyle().Foreground(gray)

	styleSelectedSuggest = lipgloss.NewStyle().Foreground(purple).Bold(true)
	styleNormalSuggest   = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff"))

	styleJSONKey   = lipgloss.NewStyle().Foreground(cyan)
	styleJSONValue = lipgloss.NewStyle().Foreground(purple)
)

type errMsg error
type resetAudioMsg struct{}

type model struct {
	state           state
	textInput       textinput.Model
	viewport        viewport.Model
	suggestions     []string
	selected        int
	userNavigated   bool
	word            string
	result          api.Response
	err             error
	historyExtended []cache.HistoryMetadata
	width           int
	height          int
	rawJSONMode     bool
	activePane      int
	audioPlaying    bool
}

func NewModel() model {
	ti := textinput.New()
	ti.Placeholder = "Search... Try filters like \":noun code\""
	ti.Focus()
	ti.CharLimit = 64

	vp := viewport.New(0, 0)

	return model{
		state:     stateInput,
		textInput: ti,
		viewport:  vp,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func fetchWordCmd(word string, filter string) tea.Cmd {
	return func() tea.Msg {
		res, err := api.Fetch(word, filter)
		if err != nil {
			return errMsg(err)
		}
		return res
	}
}

func resetAudioIndicatorCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*800, func(t time.Time) tea.Msg {
		return resetAudioMsg{}
	})
}

func buildClipboardContent(entry api.WordEntry) string {
	var out strings.Builder
	out.WriteString(fmt.Sprintf("WORD: %s\n", strings.ToUpper(entry.Word)))
	if entry.Phonetic != "" {
		out.WriteString(fmt.Sprintf("Pronunciation: %s\n", entry.Phonetic))
	}
	out.WriteString("\n")

	count := 1
	for _, meaning := range entry.Meanings {
		out.WriteString(fmt.Sprintf("[%s]\n", strings.ToUpper(meaning.PartOfSpeech)))
		for _, def := range meaning.Definitions {
			out.WriteString(fmt.Sprintf("%d. %s\n", count, def.Definition))
			count++
		}
		out.WriteString("\n")
	}
	return strings.TrimSpace(out.String())
}

func colorizeJSON(input string) string {
	var out strings.Builder
	lines := strings.Split(input, "\n")
	for _, line := range lines {
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			out.WriteString(styleJSONKey.Render(parts[0]) + ":" + styleJSONValue.Render(parts[1]) + "\n")
		} else {
			out.WriteString(line + "\n")
		}
	}
	return out.String()
}

func hasAudioTrack(res api.Response) (string, bool) {
	if len(res) == 0 {
		return "", false
	}
	for _, ph := range res[0].Phonetics {
		if ph.Audio != "" {
			return ph.Audio, true
		}
	}
	return "", false
}

func parseInputQuery(raw string) (string, string) {
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, ":") {
		parts := strings.SplitN(raw, " ", 2)
		if len(parts) == 2 {
			return parts[1], strings.TrimPrefix(parts[0], ":")
		}
	}
	return raw, ""
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		paneWidth := (m.width / 2) - 4
		paneHeight := m.height - 4

		m.textInput.Width = paneWidth - 4
		m.viewport.Width = paneWidth - 4
		m.viewport.Height = paneHeight - 6

		if len(m.result) > 0 {
			m.renderContentPane()
		}
		return m, nil

	case errMsg:
		m.err = msg
		m.state = stateResult
		m.viewport.SetContent(lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5555")).Render(fmt.Sprintf("Error Trace: %v", m.err)))
		return m, nil

	case api.Response:
		m.result = msg
		m.err = nil
		m.state = stateResult
		cache.SaveWord(m.word, msg)
		cache.SaveHistory(m.word)
		m.renderContentPane()
		return m, nil

	case resetAudioMsg:
		m.audioPlaying = false
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "q":
			if m.state != stateInput {
				m.state = stateInput
				m.activePane = 0
				m.textInput.Focus()
				return m, nil
			}

		case "tab":
			if m.state == stateInput {
				m.state = stateHistory
				m.historyExtended = cache.LoadHistoryExtended()
			} else if m.state == stateHistory {
				m.state = stateInput
				m.textInput.Focus()
			} else if m.state == stateResult {
				if m.activePane == 0 {
					m.activePane = 1
					m.textInput.Blur()
				} else {
					m.activePane = 0
					m.textInput.Focus()
				}
			}
			return m, nil

		case "up", "k":
			if m.state == stateInput && m.activePane == 0 {
				if m.selected > 0 {
					m.selected--
					m.userNavigated = true
				}
			} else if m.state == stateResult && m.activePane == 1 {
				m.viewport.LineUp(1)
			} else if m.state == stateHistory {
				if m.selected > 0 {
					m.selected--
				}
			}

		case "down", "j":
			if m.state == stateInput && m.activePane == 0 {
				if m.selected < len(m.suggestions)-1 {
					m.selected++
					m.userNavigated = true
				}
			} else if m.state == stateResult && m.activePane == 1 {
				m.viewport.LineDown(1)
			} else if m.state == stateHistory {
				if m.selected < len(m.historyExtended)-1 {
					m.selected++
				}
			}

		case "v":
			if m.state == stateResult {
				m.rawJSONMode = !m.rawJSONMode
				m.renderContentPane()
				return m, nil
			}

		case "p":
			if m.state == stateResult {
				m.audioPlaying = true
				audioURL, _ := hasAudioTrack(m.result)
				audio.PlayAudioFromURL(audioURL, m.word)
				return m, resetAudioIndicatorCmd()
			}

		case "enter":
			switch m.state {
			case stateInput:
				rawQuery := strings.TrimSpace(m.textInput.Value())
				searchWord, categoryFilter := parseInputQuery(rawQuery)

				if m.userNavigated && len(m.suggestions) > 0 && m.selected < len(m.suggestions) {
					searchWord = m.suggestions[m.selected]
				}

				if searchWord == "" {
					return m, nil
				}

				m.word = searchWord

				if categoryFilter == "" {
					if cachedRes, found := cache.LoadWord(searchWord); found {
						m.result = cachedRes
						m.err = nil
						m.state = stateResult
						cache.SaveHistory(searchWord)
						m.renderContentPane()
						return m, nil
					}
				}

				m.textInput.Blur()
				return m, fetchWordCmd(searchWord, categoryFilter)

			case stateHistory:
				if len(m.historyExtended) > 0 && m.selected < len(m.historyExtended) {
					m.word = m.historyExtended[m.selected].Word
					if cachedRes, found := cache.LoadWord(m.word); found {
						m.result = cachedRes
						m.err = nil
						m.state = stateResult
						m.renderContentPane()
					}
				}

			case stateResult:
				m.state = stateInput
				m.activePane = 0
				m.rawJSONMode = false
				m.userNavigated = false
				m.audioPlaying = false
				m.textInput.Focus()
				m.textInput.SetValue("")
				m.suggestions = nil
				m.selected = 0
			}

		case "c":
			if m.state == stateResult && len(m.result) > 0 {
				fullText := buildClipboardContent(m.result[0])
				_ = clipboard.WriteAll(fullText)
			}
		}
	}

	if m.state == stateInput && m.activePane == 0 {
		oldVal := m.textInput.Value()
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)

		cleanWord, _ := parseInputQuery(m.textInput.Value())
		if m.textInput.Value() != oldVal {
			m.suggestions = spell.Suggest(cleanWord)
			m.selected = 0
			m.userNavigated = false
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *model) renderContentPane() {
	if len(m.result) == 0 {
		return
	}

	paneWidth := (m.width / 2) - 6

	if m.rawJSONMode {
		rawBytes, _ := json.MarshalIndent(m.result, "", "  ")
		wrappedJSON := lipgloss.NewStyle().Width(paneWidth).Render(colorizeJSON(string(rawBytes)))
		m.viewport.SetContent(wrappedJSON)
		return
	}

	var rightSide strings.Builder
	entry := m.result[0]

	wordHeading := lipgloss.NewStyle().Foreground(purple).Bold(true).Render(strings.ToUpper(entry.Word))
	rightSide.WriteString(fmt.Sprintf("%s  %s\n\n", wordHeading, styleDim.Render(entry.Phonetic)))

	count := 1
	for _, meaning := range entry.Meanings {
		rightSide.WriteString(lipgloss.NewStyle().Foreground(cyan).Italic(true).Render(fmt.Sprintf("[%s]", meaning.PartOfSpeech)) + "\n")
		for _, def := range meaning.Definitions {
			defBlock := lipgloss.NewStyle().Width(paneWidth - 4).Render(fmt.Sprintf("%d. %s", count, def.Definition))
			rightSide.WriteString(defBlock + "\n")

			if len(def.Synonyms) > 0 {
				syns := lipgloss.NewStyle().Foreground(purple).Render(fmt.Sprintf("   * Synonyms: %s", strings.Join(def.Synonyms, ", ")))
				rightSide.WriteString(syns + "\n")
			}
			count++
		}
		rightSide.WriteString("\n")
	}

	m.viewport.SetContent(rightSide.String())
}

func (m model) View() string {
	if m.width < 30 || m.height < 10 {
		return "Terminal layout size too small."
	}

	paneWidth := (m.width / 2) - 2
	paneHeight := m.height - 3

	switch m.state {
	case stateHistory:
		var b strings.Builder
		b.WriteString(styleTitle.Render(" ARCHIVE HISTORY METADATA ") + "\n\n")
		if len(m.historyExtended) == 0 {
			b.WriteString(styleDim.Render("(Archive log entries trace clear)\n"))
		} else {
			for i, h := range m.historyExtended {
				if i >= paneHeight-4 {
					break
				}
				prefix := "  "
				if i == m.selected {
					prefix = "- "
				}
				line := fmt.Sprintf("%s%-16s Lookups: %-4d Last: %s", prefix, h.Word, h.QueryCount, h.Timestamp.Format("15:04:05"))
				if i == m.selected {
					b.WriteString(styleSelectedSuggest.Render(line) + "\n")
				} else {
					b.WriteString(styleNormalSuggest.Render(line) + "\n")
				}
			}
		}
		b.WriteString(styleDim.Render("\nPress [TAB] to switch back | [Enter] to pull cached record"))
		return styleActivePane.Width(m.width - 4).Height(paneHeight).Render(b.String())

	case stateInput, stateResult:
		var leftSide strings.Builder
		leftSide.WriteString(styleTitle.Render(" FDicM ") + "\n\n")
		leftSide.WriteString("Type your query below:\n")
		leftSide.WriteString(m.textInput.View() + "\n\n")
		leftSide.WriteString(styleAccent.Render("AUTOCOMPLETE INDEX") + "\n\n")

		if len(m.suggestions) == 0 {
			leftSide.WriteString(styleDim.Render("  (No direct fuzzy matches)\n"))
		} else {
			for i, s := range m.suggestions {
				if i == m.selected && m.userNavigated {
					leftSide.WriteString(styleSelectedSuggest.Render("- "+s) + "\n")
				} else {
					leftSide.WriteString(styleNormalSuggest.Render("  "+s) + "\n")
				}
			}
		}

		var rightSideRender string
		if m.state == stateInput {
			emptyMsg := lipgloss.NewStyle().Width(paneWidth - 4).Render("Waiting for query input stream execution token...")
			rightSideRender = stylePane.Width(paneWidth).Height(paneHeight).Render(styleDim.Render(emptyMsg))
		} else {
			var footerMenu strings.Builder
			if m.rawJSONMode {
				footerMenu.WriteString(styleDim.Render("[v] Regular Text | [c] Copy | [Enter] Clear"))
			} else {
				if m.audioPlaying {
					footerMenu.WriteString(lipgloss.NewStyle().Foreground(purple).Bold(true).Render("Speech Synthesis Engine Processing..."))
				} else {
					footerMenu.WriteString(lipgloss.NewStyle().Foreground(cyan).Bold(true).Render("[p] Play Audio / TTS"))
				}
				footerMenu.WriteString(styleDim.Render(" | [v] JSON | [c] Copy | [TAB] Toggle focus (Vim navigation activated)"))
			}

			content := m.viewport.View() + "\n" + footerMenu.String()

			if m.activePane == 1 {
				rightSideRender = styleActivePane.Width(paneWidth).Height(paneHeight).Render(content)
			} else {
				rightSideRender = stylePane.Width(paneWidth).Height(paneHeight).Render(content)
			}
		}

		leftStyle := stylePane
		if m.activePane == 0 {
			leftStyle = styleActivePane
		}
		leftRender := leftStyle.Width(paneWidth).Height(paneHeight).Render(leftSide.String())

		return lipgloss.JoinHorizontal(lipgloss.Top, leftRender, rightSideRender)
	}
	return ""
}
