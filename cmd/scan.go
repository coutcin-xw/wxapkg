package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"wxapkg/util"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:     "scan",
	Short:   "Scan the wechat mini program",
	Example: "  " + programName + " scan -r \"D:\\WeChat Files\\Applet\\wx12345678901234\"",
	Run: func(cmd *cobra.Command, args []string) {
		root, _ := cmd.Flags().GetString("root")
		isNew, err := cmd.Flags().GetBool("isNew")
		if err != nil {
			color.Red("%v", err)
			return
		}

		var regAppId = regexp.MustCompile(`(wx[0-9a-f]{16})`)

		var files []os.DirEntry
		if files, err = os.ReadDir(root); err != nil {
			color.Red("%v", err)
			return
		}

		var wxidInfos = make([]util.WxidInfo, 0, len(files))
		for _, file := range files {
			if !file.IsDir() || !regAppId.MatchString(file.Name()) {
				continue
			}

			var wxid = regAppId.FindStringSubmatch(file.Name())[1]
			info, err := util.WxidQuery.Query(wxid)
			info.Location = filepath.Join(root, file.Name())
			info.Wxid = wxid
			if err != nil {
				info.Error = fmt.Sprintf("%v", err)
			}

			wxidInfos = append(wxidInfos, info)
		}

		var tui = newScanTui(wxidInfos)
		if _, err := tea.NewProgram(tui, tea.WithAltScreen()).Run(); err != nil {
			color.Red("Error running program: %v", err)
			os.Exit(1)
		}

		if tui.selected == nil {
			return
		}

		output := tui.selected.Wxid
		_ = unpackCmd.Flags().Set("root", tui.selected.Location)
		_ = unpackCmd.Flags().Set("output", output)
		_ = unpackCmd.Flags().Set("isNew", strconv.FormatBool(isNew))
		detailFilePath := filepath.Join(output, "detail.json")
		unpackCmd.Run(unpackCmd, []string{"detailFilePath", detailFilePath})
		_ = os.WriteFile(detailFilePath, []byte(tui.selected.Json()), 0600)
	},
}

func init() {
	RootCmd.AddCommand(scanCmd)

	var homeDir, _ = os.UserHomeDir()
	var defaultRoot = filepath.Join(homeDir, "Documents/WeChat Files/Applet")

	scanCmd.Flags().StringP("root", "r", defaultRoot, "the mini app path")
	scanCmd.Flags().BoolP("isNew", "i", false, "is WeChat 4.0 +")
}
