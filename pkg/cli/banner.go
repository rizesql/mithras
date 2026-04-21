//nolint:lll
package cli

import (
	"github.com/fatih/color"
)

// PrintBanner prints the Mithras application logo and tagline.
func (o *Output) PrintBanner() {
	c := color.New(color.FgHiBlue).SprintFunc()
	title := color.New(color.FgWhite, color.Bold).SprintFunc()
	subtle := color.New(color.FgHiBlack).SprintFunc()

	o.Raw("")
	o.Raw("  %s    %s", c(" ▇▇      ▇▇ "), title("M I T H R A S"))
	o.Raw("  %s    %s", c(" ▇▇▇▇  ▇▇▇▇ "), title("Identity Provider"))
	o.Raw("  %s      ", c(" ▇▇▇▇▇▇▇▇▇▇ "))
	o.Raw("  %s    %s", c(" ▇▇  ▇▇  ▇▇ "), subtle("Self-contained authentication and authorization"))
	o.Raw("  %s    %s", c("▇▇▇▇    ▇▇▇▇"), subtle("with JWS tokens, audit logging, and rate limiting."))
	o.Raw("")
}

// PrintBanner prints the application banner using the default output.
func PrintBanner() {
	defaultOutput.PrintBanner()
}
