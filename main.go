package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/color"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
	"unicode"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ChatMessage represents a single message in the conversation
type ChatMessage struct {
	Sender  string `json:"sender"`
	Content string `json:"content"`
	Time    int64  `json:"time"` // Unix timestamp
}

// ChatHistory represents the entire conversation history
type ChatHistory struct {
	Messages []ChatMessage `json:"messages"`
}

// File path for saving conversation history
const historyFilePath = "goku_chat_history.json"

// saveHistory saves the chat history to a JSON file
func saveHistory(messages []ChatMessage) error {
	history := ChatHistory{
		Messages: messages,
	}

	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	err = ioutil.WriteFile(historyFilePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write history file: %w", err)
	}

	return nil
}

// loadHistory loads the chat history from a JSON file
func loadHistory() ([]ChatMessage, error) {
	// Check if file exists
	if _, err := os.Stat(historyFilePath); os.IsNotExist(err) {
		// If file doesn't exist, return empty history
		return []ChatMessage{}, nil
	}

	data, err := ioutil.ReadFile(historyFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read history file: %w", err)
	}

	var history ChatHistory
	err = json.Unmarshal(data, &history)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal history: %w", err)
	}

	return history.Messages, nil
}

// formatChatHistoryText formats the chat history for display
func formatChatHistoryText(messages []ChatMessage) string {
	var builder strings.Builder

	for _, msg := range messages {
		builder.WriteString(msg.Sender + ": " + msg.Content + "\n\n")
	}

	return builder.String()
}

// addMessageToHistory adds a new message to the history and saves it
func addMessageToHistory(messages *[]ChatMessage, sender, content string) error {
	// Create new message
	newMessage := ChatMessage{
		Sender:  sender,
		Content: content,
		Time:    time.Now().Unix(),
	}

	// Add to messages
	*messages = append(*messages, newMessage)

	// Save to file
	return saveHistory(*messages)
}

// GokuTheme is a custom theme with Dragon Ball inspired colors
type GokuTheme struct {
	fyne.Theme
}

// Colors matching Goku's outfit and energy
var (
	gokuOrange          = color.NRGBA{R: 255, G: 121, B: 0, A: 255}   // Main gi color
	gokuBlue            = color.NRGBA{R: 0, G: 162, B: 232, A: 255}   // Undershirt color
	gokuYellow          = color.NRGBA{R: 255, G: 203, B: 5, A: 255}   // Super Saiyan aura
	gokuRed             = color.NRGBA{R: 230, G: 41, B: 55, A: 255}   // Power pole
	darkBackground      = color.NRGBA{R: 20, G: 20, B: 30, A: 255}    // Dark background
	dragonBallTextColor = color.NRGBA{R: 255, G: 236, B: 214, A: 255} // Light text color
)

// NewGokuTheme creates a new Dragon Ball inspired theme
func NewGokuTheme() fyne.Theme {
	return &GokuTheme{Theme: theme.DarkTheme()}
}

// Color returns the color for the specified ThemeColorName
func (g *GokuTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNamePrimary:
		return gokuOrange
	case theme.ColorNameForeground:
		return dragonBallTextColor
	case theme.ColorNameBackground:
		return darkBackground
	case theme.ColorNameButton:
		return gokuBlue
	case theme.ColorNameFocus:
		return gokuYellow
	case theme.ColorNameHover:
		return color.NRGBA{R: 255, G: 150, B: 0, A: 30}
	case theme.ColorNameSelection:
		return gokuBlue
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 80, G: 80, B: 80, A: 255}
	case theme.ColorNameError:
		return gokuRed
	case theme.ColorNameSuccess:
		return color.NRGBA{R: 0, G: 200, B: 0, A: 255}
	case theme.ColorNamePlaceHolder:
		return color.NRGBA{R: 200, G: 200, B: 200, A: 128}
	default:
		return g.Theme.Color(name, variant)
	}
}

// Typing animation configuration
var (
	// Basic typing timing
	minTypingDelay = 40  // Minimum delay between characters in milliseconds
	maxTypingDelay = 150 // Maximum delay between characters in milliseconds

	// Variance for common keys (will type faster)
	commonLettersSpeed = 0.7 // Multiplier for common letters (e,a,i,o,t,n,s,r)
	commonLetters      = "eaiotnsr"

	// Variance for tricky keys (will type slower)
	trickeyKeysSpeed = 1.5 // Multiplier for keys that are harder to reach
	trickeyKeys      = "qzxjkwvp"

	// Pauses - thinking and phrasing
	chanceForPause   = 8                                          // 1 in X chance for a longer pause (thinking pause)
	pauseDuration    = 350                                        // Duration of normal thinking pause in milliseconds
	longThinkChance  = 25                                         // 1 in X chance for a very long thinking pause
	longThinkDelay   = 1400                                       // Duration of a long thinking pause in milliseconds
	thinkingTexts    = []string{"...", "hmm", "uh", "*thinking*"} // Texts to show during longer thinking
	sentenceEndPause = 700                                        // Extra pause after completing a sentence

	// Mistakes and corrections
	chanceForMistake     = 18               // 1 in X chance for a typing mistake
	mistakeNeighborChars = map[rune][]rune{ // Common mistaken neighboring keys
		'a': {'s', 'q', 'z'},
		's': {'a', 'd', 'w'},
		'd': {'s', 'f', 'e'},
		'f': {'d', 'g', 'r'},
		'g': {'f', 'h', 't'},
		'h': {'g', 'j', 'y'},
		'j': {'h', 'k', 'u'},
		'k': {'j', 'l', 'i'},
		'l': {'k', ';', 'o'},
		'z': {'x', 'a'},
		'x': {'z', 'c', 's'},
		'c': {'x', 'v', 'd'},
		'v': {'c', 'b', 'f'},
		'b': {'v', 'n', 'g'},
		'n': {'b', 'm', 'h'},
		'm': {'n', ',', 'j'},
	}
	chanceForDoubleError = 4   // 1 in X chance that the person makes a second error while fixing
	backspaceDelay       = 180 // Delay before backspace in milliseconds
	backspaceMinDelay    = 120 // Min delay between multiple backspaces (ms)
	backspaceMaxDelay    = 200 // Max delay between multiple backspaces (ms)

	// Goku-specific quirks
	excitedPhrases         = []string{"awesome", "power", "training", "fight", "strong", "food", "eat"}
	excitedTypingSpeedMult = 0.6  // Types faster when excited (lower is faster)
	hungrySlowdownChance   = 15   // 1 in X chance he gets distracted by food thoughts
	hungrySlowdownDelay    = 2000 // How long he pauses when thinking about food
)

// OllamaRequest represents the request structure for Ollama API
type OllamaRequest struct {
	Model    string `json:"model"`
	Prompt   string `json:"prompt"`
	System   string `json:"system"`
	Template string `json:"template"`
}

// OllamaResponse represents the response structure from Ollama API
type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

const gokuPersonality = `You are Goku from Dragon Ball. Respond in Goku's cheerful, friendly, and battle-loving personality.
Key traits to incorporate:
- Use phrases like "Hey there!", "Awesome!", "That sounds like fun!"
- Show enthusiasm for challenges and training
- Be friendly and optimistic
- Occasionally mention food, especially when you're hungry
- Use Goku's casual speech pattern
- Reference martial arts and training
- Show interest in becoming stronger
- Mention your friends like Krillin, Gohan, or Vegeta when relevant
- Use phrases like "Wow!", "Alright!", and "Let's do this!"
- Express excitement about fighting strong opponents
- Be pure-hearted and always willing to help others
Start your first message with: "Hey there! I'm Goku! *puts hand behind head and grins* What's up?"
`

// CustomEntry extends widget.Entry to handle Enter vs Shift+Enter
type CustomEntry struct {
	widget.Entry
	onEnter func()
}

// NewCustomEntry creates a new custom entry
func NewCustomEntry(onEnter func()) *CustomEntry {
	entry := &CustomEntry{onEnter: onEnter}
	entry.MultiLine = true
	entry.Wrapping = fyne.TextWrapWord
	entry.ExtendBaseWidget(entry)
	return entry
}

// TypedKey handles key events
func (e *CustomEntry) TypedKey(key *fyne.KeyEvent) {
	if key.Name == fyne.KeyReturn {
		// Always trigger the callback for Enter key
		e.onEnter()
		return
	}

	e.Entry.TypedKey(key)
}

// simulateTyping gradually displays the message character by character to simulate typing
func simulateTyping(chatHistory *widget.Label, prefix string, fullMessage string, scrollContainer *container.Scroll, onComplete func()) {
	messageChars := []rune(fullMessage)
	currentMessage := ""
	currentIndex := 0
	showingThinking := false

	// Calculate word boundaries for natural typing rhythm
	words := strings.Fields(fullMessage)
	wordEndIndexes := make([]int, 0, len(words))
	charCount := 0
	for _, word := range words {
		charCount += len(word)
		wordEndIndexes = append(wordEndIndexes, charCount)
		charCount++ // Account for the space
	}

	// Start with just the prefix
	currentText := chatHistory.Text
	chatHistory.SetText(currentText + prefix)
	scrollContainer.ScrollToBottom()

	var addNextChar func()

	// Function to add the next character
	addNextChar = func() {
		// Finished typing case
		if currentIndex >= len(messageChars) {
			// Finished typing, add newlines and call completion handler
			chatHistory.SetText(chatHistory.Text + "\n\n")
			scrollContainer.ScrollToBottom()
			onComplete()
			return
		}

		// If we are showing a thinking indicator, remove it first
		if showingThinking {
			// Remove the thinking indicator
			chatHistory.SetText(currentText + prefix + currentMessage)
			scrollContainer.ScrollToBottom()
			showingThinking = false
		}

		// Decide if we make a mistake - higher chance after word boundaries and for tricky keys
		makeMistake := rand.Intn(chanceForMistake) == 0 && currentIndex < len(messageChars)-1

		// Increase mistake chance for tricky keys
		currentChar := messageChars[currentIndex]
		if strings.ContainsRune(trickeyKeys, unicode.ToLower(currentChar)) {
			makeMistake = makeMistake || rand.Intn(chanceForMistake/2) == 0
		}

		if makeMistake {
			// Choose a wrong character - preferably adjacent on keyboard if defined
			var wrongChar rune

			// Try to use a neighboring key if available for more realistic mistakes
			if neighbors, exists := mistakeNeighborChars[unicode.ToLower(currentChar)]; exists && len(neighbors) > 0 {
				wrongChar = neighbors[rand.Intn(len(neighbors))]
				// Maintain capitalization if needed
				if unicode.IsUpper(currentChar) {
					wrongChar = unicode.ToUpper(wrongChar)
				}
			} else {
				// Random error if no neighbors defined
				wrongChar = rune(rand.Intn(26) + 'a')
				if unicode.IsUpper(currentChar) {
					wrongChar = unicode.ToUpper(wrongChar)
				}
			}

			// Add the wrong character
			chatHistory.SetText(currentText + prefix + currentMessage + string(wrongChar))
			scrollContainer.ScrollToBottom()

			// Schedule the backspace - humans notice mistakes quickly
			time.AfterFunc(time.Duration(backspaceDelay)*time.Millisecond, func() {
				// Remove the wrong character (simulate backspace)
				chatHistory.SetText(currentText + prefix + currentMessage)
				scrollContainer.ScrollToBottom()

				// Slight chance for a double-error while correcting
				if rand.Intn(chanceForDoubleError) == 0 {
					// Make another wrong keystroke before getting it right
					mistakeChar := rune(rand.Intn(26) + 'a')
					if unicode.IsUpper(currentChar) {
						mistakeChar = unicode.ToUpper(mistakeChar)
					}

					// Show the second mistake
					chatHistory.SetText(currentText + prefix + currentMessage + string(mistakeChar))
					scrollContainer.ScrollToBottom()

					// And then backspace again
					backspaceRetryDelay := rand.Intn(backspaceMaxDelay-backspaceMinDelay) + backspaceMinDelay
					time.AfterFunc(time.Duration(backspaceRetryDelay)*time.Millisecond, func() {
						chatHistory.SetText(currentText + prefix + currentMessage)
						scrollContainer.ScrollToBottom()

						// Finally schedule the correct character
						time.AfterFunc(time.Duration(minTypingDelay)*time.Millisecond, addNextChar)
					})
					return
				}

				// Schedule the correct character after a normal pause
				time.AfterFunc(time.Duration(minTypingDelay)*time.Millisecond, addNextChar)
			})
			return
		}

		// Normal case - add the next character
		currentMessage += string(messageChars[currentIndex])
		chatHistory.SetText(currentText + prefix + currentMessage)
		scrollContainer.ScrollToBottom()
		currentIndex++

		// Calculate next delay based on character type and context
		delay := rand.Intn(maxTypingDelay-minTypingDelay) + minTypingDelay

		// Type common letters faster
		if strings.ContainsRune(commonLetters, unicode.ToLower(currentChar)) {
			delay = int(float64(delay) * commonLettersSpeed)
		}

		// Type tricky letters slower
		if strings.ContainsRune(trickeyKeys, unicode.ToLower(currentChar)) {
			delay = int(float64(delay) * trickeyKeysSpeed)
		}

		// Check if we're at a word boundary (slightly longer pause)
		for _, wordEnd := range wordEndIndexes {
			if currentIndex == wordEnd {
				delay += rand.Intn(80) + 40 // Add 40-120ms pause at word boundaries
				break
			}
		}

		// Check if we just typed an excited word (Goku types faster when excited)
		for _, phrase := range excitedPhrases {
			if currentIndex >= len(phrase) &&
				strings.ToLower(string(messageChars[currentIndex-len(phrase):currentIndex])) == phrase {
				delay = int(float64(delay) * excitedTypingSpeedMult)
				break
			}
		}

		// Random chance for a "thinking" pause
		if rand.Intn(chanceForPause) == 0 && currentIndex < len(messageChars) {
			// Add a longer pause if previous character was punctuation or at end of sentence
			if currentIndex > 0 && strings.ContainsRune(",.!?", messageChars[currentIndex-1]) {
				// Longer thinking pause
				delay = pauseDuration

				// Add extra pause at end of sentences
				if strings.ContainsRune(".!?", messageChars[currentIndex-1]) {
					delay += sentenceEndPause
				}

				// Check if we should do a very long thinking pause with visible indicator
				if rand.Intn(longThinkChance) == 0 {
					// Show thinking indicator
					thinkingText := thinkingTexts[rand.Intn(len(thinkingTexts))]
					chatHistory.SetText(currentText + prefix + currentMessage + " " + thinkingText)
					scrollContainer.ScrollToBottom()
					showingThinking = true
					delay = longThinkDelay
				}

				// Special case: Goku gets distracted thinking about food
				if rand.Intn(hungrySlowdownChance) == 0 {
					thinkingText := "*stomach growls*"
					chatHistory.SetText(currentText + prefix + currentMessage + " " + thinkingText)
					scrollContainer.ScrollToBottom()
					showingThinking = true
					delay = hungrySlowdownDelay
				}
			}
		}

		// Schedule the next character
		time.AfterFunc(time.Duration(delay)*time.Millisecond, addNextChar)
	}

	// Start the animation
	addNextChar()
}

// getAIResponse sends a request to Ollama and returns the response
func getAIResponse(userMessage string) (string, error) {
	url := "http://localhost:11434/api/generate"

	// Prepare the request
	reqBody := OllamaRequest{
		Model:  "llama3.2",
		Prompt: userMessage,
		System: gokuPersonality,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %v", err)
	}

	// Send the request
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	// Read and process the response stream
	decoder := json.NewDecoder(resp.Body)
	var fullResponse strings.Builder

	for {
		var response OllamaResponse
		if err := decoder.Decode(&response); err != nil {
			if err == io.EOF {
				break
			}
			return "", fmt.Errorf("error decoding response: %v", err)
		}

		fullResponse.WriteString(response.Response)

		if response.Done {
			break
		}
	}

	return fullResponse.String(), nil
}

func main() {
	// Initialize random seed for typing animation
	rand.Seed(time.Now().UnixNano())

	// Create a new application with Goku theme
	myApp := app.New()
	myApp.Settings().SetTheme(NewGokuTheme())

	// Create a new window
	window := myApp.NewWindow("Chat with Goku")
	window.SetIcon(theme.InfoIcon()) // Use a default icon

	// Load chat history
	chatMessages, err := loadHistory()
	if err != nil {
		log.Printf("Error loading chat history: %v", err)
		// Continue with empty history
		chatMessages = []ChatMessage{}
	}

	// If history is empty, add a welcome message
	if len(chatMessages) == 0 {
		welcomeMsg := ChatMessage{
			Sender:  "Goku",
			Content: "Hey there! I'm Goku! *puts hand behind head and grins* What's up?",
			Time:    time.Now().Unix(),
		}
		chatMessages = append(chatMessages, welcomeMsg)
		// Save the initial history
		if err := saveHistory(chatMessages); err != nil {
			log.Printf("Error saving initial chat history: %v", err)
		}
	}

	// Create a chat history area with word wrapping
	chatHistory := widget.NewLabel(formatChatHistoryText(chatMessages))
	chatHistory.Wrapping = fyne.TextWrapWord

	// Create a scroll container for chat history with proper scaling
	historyBox := container.NewVBox(chatHistory)
	scrollContainer := container.NewScroll(historyBox)

	// Scroll to bottom after initial load
	scrollContainer.ScrollToBottom()

	// Create a stylized header
	headerText := canvas.NewText("CHAT WITH GOKU", gokuOrange)
	headerText.TextSize = 22
	headerText.TextStyle = fyne.TextStyle{Bold: true}
	headerText.Alignment = fyne.TextAlignCenter

	// Create a divider
	divider := canvas.NewRectangle(gokuOrange)
	divider.SetMinSize(fyne.NewSize(600, 2))

	// Create a header container
	header := container.NewVBox(
		container.NewPadded(headerText),
		divider,
	)

	// Wrap scroll container in a max size container to ensure it expands properly
	chatArea := container.NewMax(
		canvas.NewRectangle(darkBackground),
		container.NewPadded(scrollContainer),
	)

	// Create loading indicator with Goku colors
	loadingIndicator := widget.NewProgressBarInfinite()
	loadingIndicator.Hide()

	// Create send button with better styling
	sendButton := widget.NewButtonWithIcon("", theme.MailSendIcon(), nil)
	sendButton.Importance = widget.HighImportance

	// Declare sendAction variable
	var sendAction func()

	// Create input field with proper styling and handling
	input := NewCustomEntry(func() {
		sendAction()
	})
	input.SetPlaceHolder("Talk to Goku...")

	// Define the send action
	sendAction = func() {
		message := input.Text
		if strings.TrimSpace(message) != "" {
			// Add user message to chat history UI
			currentText := chatHistory.Text
			chatHistory.SetText(currentText + "You: " + message + "\n\n")

			// Add to persistent history
			if err := addMessageToHistory(&chatMessages, "You", message); err != nil {
				log.Printf("Error saving user message: %v", err)
			}

			// Scroll to the bottom immediately after adding user message
			scrollContainer.ScrollToBottom()

			// Clear input and show loading
			input.SetText("")
			loadingIndicator.Show()
			sendButton.Disable()

			// Get AI response in a goroutine
			go func() {
				response, err := getAIResponse(message)
				if err != nil {
					response = "Oops! Looks like I messed up there! *scratches head* " + err.Error()

					// Update UI but don't save error messages to history
					window.Canvas().Refresh(chatHistory)
					simulateTyping(chatHistory, "Goku: ", response, scrollContainer, func() {
						loadingIndicator.Hide()
						sendButton.Enable()
					})
				} else {
					// Only add to persistent history if it's not an error message
					addMessageToHistory(&chatMessages, "Goku", response)

					// Update UI in the main thread with typing animation
					window.Canvas().Refresh(chatHistory)
					simulateTyping(chatHistory, "Goku: ", response, scrollContainer, func() {
						loadingIndicator.Hide()
						sendButton.Enable()
					})
				}
			}()
		}
	}

	// Set the send button action
	sendButton.OnTapped = sendAction

	// Create a better styled input area with loading indicator
	inputBg := canvas.NewRectangle(color.NRGBA{R: 40, G: 40, B: 60, A: 255})
	inputArea := container.NewMax(
		inputBg,
		container.NewPadded(
			container.NewBorder(
				loadingIndicator,
				nil, nil,
				container.NewPadded(sendButton),
				input,
			),
		),
	)
	inputArea.Resize(fyne.NewSize(0, 60))

	// Create the main layout with proper spacing and organization
	mainContent := container.NewBorder(
		header, // Add header at the top
		inputArea,
		nil, nil,
		chatArea,
	)

	// Set content and handle window sizing
	window.SetContent(mainContent)

	// Set minimum size constraints
	window.Resize(fyne.NewSize(600, 400))
	window.SetFixedSize(false)

	// Make window responsive to screen size
	screen := window.Canvas().Size()
	idealWidth := screen.Width * 0.7
	idealHeight := screen.Height * 0.8

	if idealWidth < 600 {
		idealWidth = 600
	}
	if idealHeight < 400 {
		idealHeight = 400
	}

	window.Resize(fyne.NewSize(idealWidth, idealHeight))
	window.CenterOnScreen()

	// Show and run the application
	window.ShowAndRun()
}
