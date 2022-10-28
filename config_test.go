package panfigure

import (
	"testing"

	"github.com/spf13/cobra"
)

var (
	oneWordOption = &CommandOptions{
		OptName: "one",
	}
	multiWordOption = &CommandOptions{
		LongOpt: "multi-word-option",
	}
)

func TestKeyForRoot(t *testing.T) {
	c := &cobra.Command{
		Use: "root",
	}

	SetRootCommand(c)
	expected := "one"
	got := keyFor(rootCmd, oneWordOption)
	if got != expected {
		t.Errorf("expected %s | got %s", expected, got)
	}
}

func TestKeyForSimple(t *testing.T) {
	c := &cobra.Command{
		Use: "top",
	}
	rootCmd.AddCommand(c)

	expected := "top.one"
	got := keyFor(c, oneWordOption)
	if got != expected {
		t.Errorf("expected %s | got %s", expected, got)
	}
}

func TestKeyForNestedOnce(t *testing.T) {
	c := &cobra.Command{
		Use: "top",
	}
	rootCmd.AddCommand(c)
	s := &cobra.Command{
		Use: "sub",
	}
	c.AddCommand(s)

	expected := "top.sub.one"
	got := keyFor(s, oneWordOption)
	if got != expected {
		t.Errorf("expected %s | got %s", expected, got)
	}
}

func TestKeyForMultiWord(t *testing.T) {
	c := &cobra.Command{
		Use: "top",
	}
	rootCmd.AddCommand(c)
	s := &cobra.Command{
		Use: "sub",
	}
	c.AddCommand(s)

	expected := "top.sub.multi_word_option"
	got := keyFor(s, multiWordOption)
	if got != expected {
		t.Errorf("expected %s | got %s", expected, got)
	}
}
