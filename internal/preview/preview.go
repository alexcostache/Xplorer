package preview

import (
	"bufio"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/lexers"
	"github.com/nsf/termbox-go"
	"golang.org/x/text/width"
)

// Manager handles file preview operations
type Manager struct {
	lastPreviewLines []string
	scrollOffset     int
}

// NewManager creates a new preview manager
func NewManager() *Manager {
	return &Manager{
		lastPreviewLines: nil,
		scrollOffset:     0,
	}
}

// GetLines returns the cached preview lines
func (m *Manager) GetLines() []string {
	return m.lastPreviewLines
}

// GetScrollOffset returns the current scroll offset
func (m *Manager) GetScrollOffset() int {
	return m.scrollOffset
}

// SetScrollOffset sets the scroll offset
func (m *Manager) SetScrollOffset(offset int) {
	m.scrollOffset = offset
}

// ScrollDown scrolls the preview down
func (m *Manager) ScrollDown(amount, visibleLines int) {
	if len(m.lastPreviewLines) > visibleLines {
		maxOffset := len(m.lastPreviewLines) - visibleLines
		m.scrollOffset = min(m.scrollOffset+amount, maxOffset)
	}
}

// ScrollUp scrolls the preview up
func (m *Manager) ScrollUp(amount int) {
	m.scrollOffset = max(m.scrollOffset-amount, 0)
}

// ResetScroll resets the scroll offset
func (m *Manager) ResetScroll() {
	m.scrollOffset = 0
}

// LoadPreview loads preview for a file or directory
func (m *Manager) LoadPreview(path string, showHidden bool, maxLines int) error {
	info, err := os.Stat(path)
	if err != nil {
		m.lastPreviewLines = []string{err.Error()}
		m.scrollOffset = 0
		return err
	}

	if info.IsDir() {
		entries, err := os.ReadDir(path)
		if err != nil {
			m.lastPreviewLines = []string{err.Error()}
			m.scrollOffset = 0
			return err
		}
		
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Name() < entries[j].Name()
		})
		
		var lines []string
		for _, entry := range entries {
			if !showHidden && strings.HasPrefix(entry.Name(), ".") {
				continue
			}
			lines = append(lines, entry.Name())
			if maxLines > 0 && len(lines) >= maxLines {
				break
			}
		}
		m.lastPreviewLines = lines
		m.scrollOffset = 0
		return nil
	}

	// Try to read text file
	file, err := os.Open(path)
	if err != nil {
		m.lastPreviewLines = []string{describeFileByExt(filepath.Base(path))}
		m.scrollOffset = 0
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		line := scanner.Text()
		
		// Detect binary files
		if strings.ContainsRune(line, '\x00') {
			m.lastPreviewLines = []string{"[" + describeFileByExt(filepath.Base(path)) + "]"}
			m.scrollOffset = 0
			return nil
		}
		
		lines = append(lines, line)
		if maxLines > 0 && len(lines) >= maxLines {
			break
		}
	}
	
	if err := scanner.Err(); err != nil {
		m.lastPreviewLines = []string{"[error reading file]"}
		m.scrollOffset = 0
		return nil
	}
	
	if len(lines) == 0 {
		m.lastPreviewLines = []string{"[" + describeFileByExt(filepath.Base(path)) + "]"}
		m.scrollOffset = 0
		return nil
	}
	
	m.lastPreviewLines = lines
	m.scrollOffset = 0
	return nil
}

// DrawText draws syntax-highlighted text with theme-aware colors
func DrawText(x, y int, line string, lang string, colorText, colorBackground, colorDim termbox.Attribute) {
	// Fallback for no language
	if lang == "" {
		for i, r := range line {
			termbox.SetCell(x+i, y, r, colorText, colorBackground)
		}
		return
	}

	lexer := lexers.Get(lang)
	if lexer == nil {
		lexer = lexers.Analyse(line)
	}
	if lexer == nil {
		for i, r := range line {
			termbox.SetCell(x+i, y, r, colorText, colorBackground)
		}
		return
	}

	// Tokenize line
	code := line + "\n"
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		for i, r := range line {
			termbox.SetCell(x+i, y, r, colorText, colorBackground)
		}
		return
	}

	xPos := x
	w, _ := termbox.Size()

	for token := iterator(); token != chroma.EOF; token = iterator() {
		fg := getSyntaxColor(token.Type, colorText, colorDim)

		for _, r := range token.Value {
			if r == '\n' || xPos >= w {
				break
			}
			termbox.SetCell(xPos, y, r, fg, colorBackground)
			xPos += RuneWidth(r)
		}
	}
}

// getSyntaxColor returns appropriate color for syntax token type
func getSyntaxColor(tokenType chroma.TokenType, colorText, colorDim termbox.Attribute) termbox.Attribute {
	// Keywords (if, for, func, class, etc.)
	if tokenType == chroma.Keyword ||
	   tokenType == chroma.KeywordConstant ||
	   tokenType == chroma.KeywordDeclaration ||
	   tokenType == chroma.KeywordNamespace ||
	   tokenType == chroma.KeywordPseudo ||
	   tokenType == chroma.KeywordReserved ||
	   tokenType == chroma.KeywordType {
		return termbox.ColorBlue
	}
	
	// Strings
	if tokenType == chroma.String ||
	   tokenType == chroma.LiteralString ||
	   tokenType == chroma.LiteralStringAffix ||
	   tokenType == chroma.LiteralStringBacktick ||
	   tokenType == chroma.LiteralStringChar ||
	   tokenType == chroma.LiteralStringDelimiter ||
	   tokenType == chroma.LiteralStringDoc ||
	   tokenType == chroma.LiteralStringDouble ||
	   tokenType == chroma.LiteralStringEscape ||
	   tokenType == chroma.LiteralStringHeredoc ||
	   tokenType == chroma.LiteralStringInterpol ||
	   tokenType == chroma.LiteralStringOther ||
	   tokenType == chroma.LiteralStringRegex ||
	   tokenType == chroma.LiteralStringSingle ||
	   tokenType == chroma.LiteralStringSymbol {
		return termbox.ColorGreen
	}
	
	// Comments
	if tokenType == chroma.Comment ||
	   tokenType == chroma.CommentHashbang ||
	   tokenType == chroma.CommentMultiline ||
	   tokenType == chroma.CommentSingle ||
	   tokenType == chroma.CommentSpecial ||
	   tokenType == chroma.CommentPreproc ||
	   tokenType == chroma.CommentPreprocFile {
		return colorDim
	}
	
	// Numbers
	if tokenType == chroma.Number ||
	   tokenType == chroma.LiteralNumber ||
	   tokenType == chroma.LiteralNumberBin ||
	   tokenType == chroma.LiteralNumberFloat ||
	   tokenType == chroma.LiteralNumberHex ||
	   tokenType == chroma.LiteralNumberInteger ||
	   tokenType == chroma.LiteralNumberIntegerLong ||
	   tokenType == chroma.LiteralNumberOct {
		return termbox.ColorYellow
	}
	
	// Functions/Methods
	if tokenType == chroma.Name ||
	   tokenType == chroma.NameFunction ||
	   tokenType == chroma.NameClass ||
	   tokenType == chroma.NameBuiltin ||
	   tokenType == chroma.NameBuiltinPseudo {
		return termbox.ColorCyan
	}
	
	// Operators
	if tokenType == chroma.Operator ||
	   tokenType == chroma.OperatorWord ||
	   tokenType == chroma.Punctuation {
		return termbox.ColorMagenta
	}
	
	// Default to text color
	return colorText
}

// DetectLanguage detects the programming language from filename
func DetectLanguage(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	languages := map[string]string{
		".go":   "go",
		".py":   "python",
		".js":   "javascript",
		".jsx":  "javascript",
		".ts":   "typescript",
		".tsx":  "typescript",
		".json": "json",
		".sh":   "shell",
		".html": "html",
		".htm":  "html",
		".css":  "css",
		".c":    "c",
		".h":    "c",
		".cpp":  "cpp",
		".hpp":  "cpp",
		".cc":   "cpp",
		".cxx":  "cpp",
		".java": "java",
		".rb":   "ruby",
		".rs":   "rust",
		".php":  "php",
	}
	
	if lang, ok := languages[ext]; ok {
		return lang
	}
	return ""
}

// RuneWidth returns the display width of a rune
func RuneWidth(r rune) int {
	prop := width.LookupRune(r)
	switch prop.Kind() {
	case width.EastAsianWide, width.EastAsianFullwidth:
		return 2
	default:
		return 1
	}
}

// describeFileByExt returns a description of a file type
func describeFileByExt(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	
	descriptions := map[string]string{
		".exe":  "EXE File",
		".dll":  "DLL File",
		".png":  "Image File",
		".jpg":  "Image File",
		".jpeg": "Image File",
		".gif":  "Image File",
		".svg":  "Image File",
		".zip":  "Archive File",
		".tar":  "Archive File",
		".gz":   "Archive File",
		".rar":  "Archive File",
		".pdf":  "PDF Document",
		".mp4":  "Video File",
		".mkv":  "Video File",
		".avi":  "Video File",
		".mp3":  "Audio File",
		".wav":  "Audio File",
		".flac": "Audio File",
		".bin":  "Binary File",
		".dat":  "Binary File",
	}
	
	if desc, ok := descriptions[ext]; ok {
		if ext != "" {
			return desc + " (" + ext + ")"
		}
		return desc
	}
	
	if ext != "" {
		return "Unknown File (" + ext + ")"
	}
	return "Unknown File"
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Made with Bob
