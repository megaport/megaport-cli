package output

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

func PrintSuccess(format string, noColor bool, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if noColor {
		fmt.Printf("✓ %s\n", msg)
	} else {
		fmt.Print(color.GreenString("✓ "))
		fmt.Println(msg)
	}
}

func FormatSuccess(msg string, noColor bool) string {
	if noColor {
		return "successfully"
	}
	return color.GreenString("successfully")
}

func PrintResourceSuccess(resourceType, action, uid string, noColor bool) {
	uidFormatted := FormatUID(uid, noColor)
	if strings.HasSuffix(action, "ed") {
		PrintSuccess("%s %s %s", noColor, resourceType, action, uidFormatted)
	} else {
		PrintSuccess("%s %sd %s", noColor, resourceType, action, uidFormatted)
	}
}

func PrintResourceCreated(resourceType, uid string, noColor bool) {
	PrintSuccess("%s created %s", noColor, resourceType, FormatUID(uid, noColor))
}

func PrintResourceUpdated(resourceType, uid string, noColor bool) {
	PrintSuccess("%s updated %s", noColor, resourceType, FormatUID(uid, noColor))
}

func PrintResourceDeleted(resourceType, uid string, immediate, noColor bool) {
	msg := fmt.Sprintf("%s deleted %s", resourceType, FormatUID(uid, noColor))
	if immediate {
		msg += "\nThe resource will be deleted immediately"
	} else {
		msg += "\nThe resource will be deleted at the end of the current billing period"
	}
	PrintSuccess(msg, noColor)
}

var spinnerChars = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type Spinner struct {
	stop      chan bool
	stopped   bool
	frameRate time.Duration
	mu        sync.Mutex
	noColor   bool
}

func NewSpinner(noColor bool) *Spinner {
	return &Spinner{
		stop:      make(chan bool),
		frameRate: 100 * time.Millisecond,
		noColor:   noColor,
	}
}

func (s *Spinner) Start(prefix string) {
	s.mu.Lock()
	if s.stopped {
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()

	go func() {
		for i := 0; ; i++ {
			select {
			case <-s.stop:
				return
			default:
				s.mu.Lock()
				if s.stopped {
					s.mu.Unlock()
					return
				}
				frame := spinnerChars[i%len(spinnerChars)]
				if s.noColor {
					fmt.Printf("\r\033[K%s %s", frame, prefix)
				} else {
					fmt.Printf("\r\033[K%s %s", color.CyanString(frame), prefix)
				}
				s.mu.Unlock()
				time.Sleep(s.frameRate)
			}
		}
	}()
}

func (s *Spinner) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stopped {
		return
	}
	s.stopped = true
	s.stop <- true
	fmt.Print("\r\033[K")
}

func (s *Spinner) StopWithSuccess(msg string) {
	s.Stop()
	if s.noColor {
		fmt.Printf("✓ %s\n", msg)
	} else {
		fmt.Print(color.GreenString("✓ "))
		fmt.Println(msg)
	}
}

func PrintResourceCreating(resourceType, uid string, noColor bool) *Spinner {
	uidFormatted := FormatUID(uid, noColor)
	msg := fmt.Sprintf("Creating %s %s...", resourceType, uidFormatted)
	spinner := NewSpinner(noColor)
	spinner.Start(msg)
	return spinner
}

func PrintResourceUpdating(resourceType, uid string, noColor bool) *Spinner {
	uidFormatted := FormatUID(uid, noColor)
	msg := fmt.Sprintf("Updating %s %s...", resourceType, uidFormatted)
	spinner := NewSpinner(noColor)
	spinner.Start(msg)
	return spinner
}

func PrintResourceDeleting(resourceType, uid string, noColor bool) *Spinner {
	uidFormatted := FormatUID(uid, noColor)
	msg := fmt.Sprintf("Deleting %s %s...", resourceType, uidFormatted)
	spinner := NewSpinner(noColor)
	spinner.Start(msg)
	return spinner
}

func PrintResourceListing(resourceType string, noColor bool) *Spinner {
	msg := fmt.Sprintf("Listing %ss...", resourceType)
	spinner := NewSpinner(noColor)
	spinner.Start(msg)
	return spinner
}

func PrintResourceGetting(resourceType, uid string, noColor bool) *Spinner {
	uidFormatted := FormatUID(uid, noColor)
	msg := fmt.Sprintf("Getting %s %s details...", resourceType, uidFormatted)
	spinner := NewSpinner(noColor)
	spinner.Start(msg)
	return spinner
}

func PrintListingResourceTags(resourceType, uid string, noColor bool) *Spinner {
	uidFormatted := FormatUID(uid, noColor)
	msg := fmt.Sprintf("Listing resource tags for %s %s...", resourceType, uidFormatted)
	spinner := NewSpinner(noColor)
	spinner.Start(msg)
	return spinner
}

func PrintResourceValidating(resourceType string, noColor bool) *Spinner {
	msg := fmt.Sprintf("Validating %s order...", resourceType)
	spinner := NewSpinner(noColor)
	spinner.Start(msg)
	return spinner
}

func PrintLoggingIn(noColor bool) *Spinner {
	msg := "Logging in to Megaport..."
	spinner := NewSpinner(noColor)
	spinner.Start(msg)
	return spinner
}

func PrintCustomSpinner(action, resourceId string, noColor bool) *Spinner {
	uidFormatted := FormatUID(resourceId, noColor)
	msg := fmt.Sprintf("%s %s...", action, uidFormatted)
	spinner := NewSpinner(noColor)
	spinner.Start(msg)
	return spinner
}

func PrintError(format string, noColor bool, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if noColor {
		fmt.Printf("✗ %s\n", msg)
	} else {
		fmt.Print(color.RedString("✗ "))
		fmt.Println(msg)
	}
}

func PrintWarning(format string, noColor bool, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if noColor {
		fmt.Printf("⚠ %s\n", msg)
	} else {
		fmt.Print(color.YellowString("⚠ "))
		fmt.Println(msg)
	}
}

func PrintInfo(format string, noColor bool, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if noColor {
		fmt.Printf("ℹ %s\n", msg)
	} else {
		fmt.Print(color.BlueString("ℹ "))
		fmt.Println(msg)
	}
}

func FormatConfirmation(msg string, noColor bool) string {
	if noColor {
		return fmt.Sprintf("%s [y/N]", msg)
	}
	return fmt.Sprintf("%s %s", msg, color.YellowString("[y/N]"))
}

func FormatPrompt(msg string, noColor bool) string {
	if noColor {
		return msg
	}
	return color.BlueString(msg)
}

func FormatExample(example string, noColor bool) string {
	if noColor {
		return example
	}
	return color.CyanString(example)
}

func FormatCommandName(name string, noColor bool) string {
	if noColor {
		return name
	}
	return color.MagentaString(name)
}

func FormatRequiredFlag(flag string, description string, noColor bool) string {
	if noColor {
		return fmt.Sprintf("%s (REQUIRED): %s", flag, description)
	}
	return fmt.Sprintf("%s: %s", color.YellowString("%s (REQUIRED)", flag), description)
}

func FormatOptionalFlag(flag string, description string, noColor bool) string {
	if noColor {
		return fmt.Sprintf("%s: %s", flag, description)
	}
	return fmt.Sprintf("%s: %s", color.BlueString(flag), description)
}

func FormatJSONExample(json string, noColor bool) string {
	if noColor {
		return json
	}
	return color.GreenString(json)
}

func FormatUID(uid string, noColor bool) string {
	if noColor {
		return uid
	}
	return color.CyanString(uid)
}

func StripANSIColors(s string) string {
	re := regexp.MustCompile("\x1b\\[[0-9;]*m")
	return re.ReplaceAllString(s, "")
}

func FormatOldValue(value string, noColor bool) string {
	if noColor {
		return value
	}
	return color.New(color.FgYellow).Sprint(value)
}

func FormatNewValue(value string, noColor bool) string {
	if noColor {
		return value
	}
	return color.New(color.FgGreen).Sprint(value)
}
