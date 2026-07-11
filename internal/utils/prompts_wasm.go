//go:build js && wasm

package utils

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/megaport/megaport-cli/internal/wasm"
)

// WASM versions of the prompt functions that use JavaScript callbacks
// instead of stdin/stdout

// promptLineBreak separates lines inside a prompt message. The host renders
// prompt messages straight to a terminal that is not in convertEol mode, so a
// bare \n would move down without returning to column 0 (a stair-step). Buffer
// output (see flush) uses a plain \n because its renderer converts EOLs.
const promptLineBreak = "\r\n"

// sanitizeForTerminal strips C0 controls, DEL, and C1 controls from text that
// is folded into a terminal-bound message. The host writes prompt messages to
// xterm without escaping, so a crafted resource name or tag value could
// otherwise inject cursor/erase sequences that rewrite the purchase summary
// shown before the [y/N]. C1 (0x80-0x9f) is stripped because some terminals
// treat it as escape controls (e.g. 0x9b as CSI). Only display copies are
// sanitized; the values stored on the order come from the prompt responses,
// which are untouched. Structural line breaks are inserted by callers after
// sanitizing, so they are preserved.
func sanitizeForTerminal(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, s)
}

func init() {
	// Override the default prompt functions with WASM-specific ones
	SetPrompt(wasmPrompt)
	SetConfirmPrompt(wasmConfirmPrompt)
	SetBuyConfirmPrompt(wasmBuyConfirmPrompt)
	SetDesignConfirmPrompt(wasmDesignConfirmPrompt)
	SetPasswordPrompt(wasmPasswordPrompt)
	SetResourcePrompt(wasmResourcePrompt)
	SetSecretResourcePrompt(wasmSecretResourcePrompt)
	SetResourceTagsPrompt(wasmResourceTagsPrompt)
	SetUpdateResourceTagsPrompt(wasmUpdateResourceTagsPrompt)
}

func wasmPrompt(msg string, noColor bool) (string, error) {
	// In WASM, we use the JavaScript prompt callback
	return wasm.PromptForInput(msg, "text", "")
}

func wasmConfirmPrompt(question string, noColor bool) bool {
	// Add [y/N] to the question for clarity
	fullQuestion := question + " [y/N]"

	response, err := wasm.PromptForInput(fullQuestion, "confirm", "")
	if err != nil {
		return false
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

// wasmBuyConfirmPrompt folds the purchase summary into the live prompt message
// so the customer sees every detail before the [y/N]. On the WASM target the
// native printConfirmSummary writes to os.Stdout, which the browser routes to
// the console (invisible to the terminal), so the summary must travel on the
// prompt channel instead.
func wasmBuyConfirmPrompt(resourceType string, details []BuyConfirmDetail, noColor bool) bool {
	msg := buildConfirmSummary("Purchase Summary:", resourceType, details, "Proceed with purchase?")
	return wasmConfirmPrompt(msg, noColor)
}

func wasmDesignConfirmPrompt(resourceType string, details []BuyConfirmDetail, noColor bool) bool {
	msg := buildConfirmSummary("Design Summary:", resourceType, details, "Proceed with creation?")
	return wasmConfirmPrompt(msg, noColor)
}

// buildConfirmSummary renders the summary block and question as a single string
// for delivery over the prompt channel: header, resource type, non-empty detail
// lines, a blank line, then the question, matching the native printConfirmSummary
// content (minus its color and leading blank line, which suit a scrolling TTY
// rather than a discrete prompt payload). Lines use \r\n because the host renders
// prompt messages on a terminal that is not in convertEol mode, so a bare \n
// would stair-step the summary. The resource type and detail values are
// sanitized: they can be user-supplied (e.g. a resource name).
func buildConfirmSummary(header, resourceType string, details []BuyConfirmDetail, question string) string {
	lines := []string{header, fmt.Sprintf("  Resource Type: %s", sanitizeForTerminal(resourceType))}
	for _, d := range details {
		if d.Value != "" {
			lines = append(lines, fmt.Sprintf("  %s: %s", d.Key, sanitizeForTerminal(d.Value)))
		}
	}
	lines = append(lines, "", question)
	return strings.Join(lines, promptLineBreak)
}

func wasmPasswordPrompt(msg string, noColor bool) (string, error) {
	return wasm.PromptForInput(msg, "password", "")
}

func wasmResourcePrompt(resourceType string, msg string, noColor bool) (string, error) {
	// Use the resource-specific prompt type
	return wasm.PromptForInput(msg, "resource", resourceType)
}

// wasmSecretResourcePrompt delegates to the JS prompt callback with a
// "password" prompt type so the host UI can mask the input (e.g. render an
// <input type="password">).
func wasmSecretResourcePrompt(resourceType string, msg string, noColor bool) (string, error) {
	return wasm.PromptForInput(msg, "password", resourceType)
}

// wasmPromptBuffer accumulates informational lines (instructions, warnings,
// per-tag echoes) and folds them into the next prompt message, so they reach
// the live prompt channel instead of os.Stdout, which the browser routes to the
// console. Trailing lines with no following prompt are flushed to the output
// buffer instead.
type wasmPromptBuffer struct {
	lines []string
}

func (b *wasmPromptBuffer) add(line string) {
	b.lines = append(b.lines, sanitizeForTerminal(line))
}

// prompt prepends any buffered lines to msg, sends it over the live prompt
// channel, and clears the buffer.
func (b *wasmPromptBuffer) prompt(msg string, noColor bool) (string, error) {
	return wasmPrompt(b.take(msg), noColor)
}

// confirm prepends any buffered lines to question, asks it over the live
// prompt channel, and clears the buffer.
func (b *wasmPromptBuffer) confirm(question string, noColor bool) bool {
	return wasmConfirmPrompt(b.take(question), noColor)
}

func (b *wasmPromptBuffer) take(msg string) string {
	msg = sanitizeForTerminal(msg)
	if len(b.lines) == 0 {
		return msg
	}
	prefix := strings.Join(b.lines, promptLineBreak)
	b.lines = nil
	return prefix + promptLineBreak + msg
}

// flush writes any remaining buffered lines (no following prompt) to the output
// buffer so they are captured rather than lost to the console.
func (b *wasmPromptBuffer) flush() {
	if len(b.lines) == 0 {
		return
	}
	_, _ = wasm.WasmOutputBuffer.Write([]byte(strings.Join(b.lines, "\n") + "\n"))
	b.lines = nil
}

func wasmResourceTagsPrompt(noColor bool) (map[string]string, error) {
	addTags := wasmConfirmPrompt("Would you like to add resource tags?", noColor)
	if !addTags {
		return nil, nil
	}

	tags := make(map[string]string)
	var buf wasmPromptBuffer

	// Inform user how to finish entering tags; shown with the first key prompt.
	buf.add("Enter tags (key and value). Enter empty key to finish.")

	for {
		key, err := buf.prompt("Tag key (empty to finish):", noColor)
		if err != nil {
			return nil, err
		}

		if key == "" {
			break
		}

		value, err := buf.prompt(fmt.Sprintf("Tag value for '%s':", key), noColor)
		if err != nil {
			return nil, err
		}

		tags[key] = value
	}

	if len(tags) > 0 {
		buf.add("Tags added:")
		for k, v := range tags {
			buf.add(fmt.Sprintf("  %s: %s", k, v))
		}
	}
	buf.flush()

	return tags, nil
}

func wasmUpdateResourceTagsPrompt(existingTags map[string]string, noColor bool) (map[string]string, error) {
	var buf wasmPromptBuffer

	// Warning about replacing tags; shown with the continue prompt.
	buf.add("⚠️  Warning: This operation will replace all existing tags with the new set of tags you define.")

	if len(existingTags) > 0 {
		buf.add("Current tags:")
		for k, v := range existingTags {
			buf.add(fmt.Sprintf("  %s: %s", k, v))
		}
	} else {
		buf.add("No existing tags found.")
	}

	proceed := buf.confirm("Do you want to continue with updating tags?", noColor)
	if !proceed {
		return nil, fmt.Errorf("tag update cancelled by user")
	}

	// With no existing tags, defer to the add flow (mirrors the native path).
	if len(existingTags) == 0 {
		return wasmResourceTagsPrompt(noColor)
	}

	tags := make(map[string]string)

	buf.add("")
	buf.add("Choose how you want to update tags:")
	buf.add("1. Start with a clean slate (remove all existing tags)")
	buf.add("2. Start with existing tags and modify them")

	choice, err := buf.prompt("Enter choice (1 or 2):", noColor)
	if err != nil {
		return nil, err
	}

	if choice == "2" {
		for k, v := range existingTags {
			tags[k] = v
		}

		buf.add("")
		buf.add("You can now modify, add, or remove tags.")
		buf.add("To remove a tag, enter its key and an empty value.")
	} else if choice != "1" {
		return nil, fmt.Errorf("invalid choice: %s", choice)
	}

	buf.add("")
	buf.add("Enter tags (key and value). Enter empty key to finish.")
	for {
		key, err := buf.prompt("Tag key (empty to finish):", noColor)
		if err != nil {
			return nil, err
		}

		if key == "" {
			break
		}

		value, err := buf.prompt(fmt.Sprintf("Tag value for '%s' (empty to remove):", key), noColor)
		if err != nil {
			return nil, err
		}

		// Empty value means remove the tag
		if value == "" && tags[key] != "" {
			delete(tags, key)
			buf.add(fmt.Sprintf("  Removed tag: %s", key))
		} else if value != "" {
			tags[key] = value
			buf.add(fmt.Sprintf("  Updated tag: %s: %s", key, value))
		}
	}

	buf.add("")
	buf.add("Final tags that will be applied:")
	if len(tags) > 0 {
		for k, v := range tags {
			buf.add(fmt.Sprintf("  %s: %s", k, v))
		}
	} else {
		buf.add("  No tags - all existing tags will be removed")
	}

	confirmApply := buf.confirm("Apply these changes?", noColor)
	if !confirmApply {
		return nil, fmt.Errorf("tag update cancelled by user")
	}

	return tags, nil
}
