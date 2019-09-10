package runner

import (
	"bufio"
	"fmt"
	"frpok/config"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"runtime"
	"strings"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/go-ini/ini"
)

// Runner frp runner
type Runner struct {
	// frpok config
	cfg *config.Config
	// runtime config file
	runtimeCfgFile *os.File
	// frpc cmd
	frpcCmd *exec.Cmd
}

// init runner
func (runner *Runner) init(cfg *config.Config) {
	runner.cfg = cfg
}

// run with ui
func (runner *Runner) runUI(topTips []string) {
	// runner.runFrpc()
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	// top port tips
	frpokTips := widgets.NewParagraph()
	frpokTips.Text = strings.Join(topTips, "\n")
	frpokTips.Title = "Frpok"
	frpokTips.TitleStyle.Fg = ui.ColorCyan

	// frp logs
	frpcLogs := widgets.NewList()
	frpcLogs.Title = "Frpc logs"
	frpcLogs.TitleStyle.Fg = ui.ColorMagenta
	frpcLogs.TextStyle.Fg = ui.ColorYellow

	grid := ui.NewGrid()
	termWidth, termHeight := ui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)

	grid.Set(ui.NewRow(1.0/3, frpokTips), ui.NewRow(2.0/3, frpcLogs))

	ui.Render(grid)

	uiEvents := ui.PollEvents()
	frpcTask := runner.sunFrpc()

	for {
		select {
		case msg := <-frpcTask:
			switch msg {
			case "EOF":
				frpcLogs.Rows = append(frpcLogs.Rows, fmt.Sprintf("[%s](fg:%s)", "Frpc not running!", "red"))
			default:
				var styledMsg = msg

				if strings.Index(msg, "[I]") != -1 {
					styledMsg = fmt.Sprintf("[%s](fg:%s)", styledMsg, "blue")
				}

				if strings.Index(msg, "[W]") != -1 {
					styledMsg = fmt.Sprintf("[%s](fg:%s)", styledMsg, "yellow")
				}

				if strings.Index(msg, "[E]") != -1 {
					styledMsg = fmt.Sprintf("[%s](fg:%s)", styledMsg, "red")
				}

				frpcLogs.Rows = append(frpcLogs.Rows, styledMsg)
				ui.Render(grid)
			}
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				runner.stopFrpc()
				return
			case "j", "<Down>", "<MouseWheelDown>":
				frpcLogs.ScrollDown()
				ui.Render(grid)
			case "k", "<Up>", "<MouseWheelUp>":
				frpcLogs.ScrollUp()
				ui.Render(grid)
			case "<Resize>":
				payload := e.Payload.(ui.Resize)
				grid.SetRect(0, 0, payload.Width, payload.Height)
				ui.Clear()
				ui.Render(grid)
			}
		}
	}

}

// stop frpc
func (runner *Runner) stopFrpc() {
	if runner.frpcCmd != nil {
		runner.frpcCmd.Process.Kill()
		runner.frpcCmd = nil
	}
}

// run frpc
func (runner *Runner) sunFrpc() <-chan string {

	ch := make(chan string)

	go func(r *Runner) {
		key := runner.cfg.Frpok.Key("frpc_path")

		var frpcPath string

		if key != nil {
			frpcPath = key.Value()
		} else {

			wd, err := os.Getwd()
			if err != nil {
				log.Fatal(err)
			}

			ext := ""

			// in windows with ext
			if runtime.GOOS == "windows" {
				ext = ".exe"
			}

			// default frpc in pwd
			frpcPath = path.Join(wd, "frpc"+ext)
		}

		r.frpcCmd = exec.Command(frpcPath, "-c", runner.runtimeCfgFile.Name())

		stdout, err := r.frpcCmd.StdoutPipe()
		if err != nil {
			log.Fatal(err)
		}

		r.frpcCmd.Start()

		outScanner := bufio.NewScanner(stdout)
		outScanner.Split(bufio.ScanLines)

		reg, err := regexp.Compile(`(\[1;\d{2}m)|(\[0[^\]])`)

		if err != nil {
			log.Fatal(err)
		}

		for outScanner.Scan() {
			m := outScanner.Text()
			ch <- reg.ReplaceAllString(m, "")
		}

		r.frpcCmd.Wait()

		ch <- "EOF"

	}(runner)

	return ch
}

// copy section
func copySection(cfg *ini.File, source *ini.Section) *ini.Section {

	section, e := cfg.NewSection(source.Name())

	if e != nil {
		log.Fatal(e)
	}

	keys := source.Keys()

	for i := range keys {
		sourceKey := keys[i]
		section.NewKey(sourceKey.Name(), sourceKey.Value())
	}

	return section
}

// writeRuntimeConfig write runtime config in tmp dir
func (runner *Runner) writeRuntimeConfig(cfg *ini.File) {

	var err error

	runner.runtimeCfgFile, err = ioutil.TempFile("", "frpok")

	if err != nil {
		log.Fatal(err)
	}

	// save to tmp file
	err = cfg.SaveTo(runner.runtimeCfgFile.Name())

	if err != nil {
		log.Fatal(err)
	}
}
