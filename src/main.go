package main

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/xxscloud5722/k9s-help/src/kube"
	"os"
	"os/user"
	"path"
	"runtime"
	"strings"
)

func main() {
	var rootCmd = &cobra.Command{
		Use: "k9s_help",
		Run: func(cmd *cobra.Command, args []string) {
			kubeConfig, err := cmd.Flags().GetString("kubeconfig")
			if err != nil {
				color.Red(err.Error())
				return
			}
			if kubeConfig == "" {
				kubeConfig, err = userHome()
				if err != nil {
					color.Red(err.Error())
					return
				}
				kubeConfig = pathBeautification(path.Join(kubeConfig, ".kube", "config"))
			}
			output, err := cmd.Flags().GetString("output")
			if err != nil {
				color.Red(err.Error())
				return
			}
			if output == "" {
				output, err = os.Getwd()
				if err != nil {
					color.Red(err.Error())
					return
				}
			}
			ignore, err := cmd.Flags().GetBool("ignore")
			if err != nil {
				color.Red(err.Error())
				return
			}
			color.Blue("[kube] Loading Config: " + kubeConfig)
			k, err := kube.New(kubeConfig, "", output)
			if err != nil {
				color.Red("[kube] Execute fail: " + err.Error())
				return
			}
			err = k.RefreshProject(ignore)
			if err != nil {
				color.Red("[kube] Execute fail: " + err.Error())
				return
			}
			color.Green("[kube] Project refresh completed !")
		},
	}
	rootCmd.PersistentFlags().String("kubeconfig", "", "Specify kubeconfig")
	rootCmd.PersistentFlags().Bool("ignore", true, "Specify kubeconfig")
	rootCmd.PersistentFlags().String("output", "", "Output Path")
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	err := rootCmd.Execute()
	if err != nil {
		return
	}
}

func userHome() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}
	return currentUser.HomeDir, nil
}

func pathBeautification(path string) string {
	if runtime.GOOS == "windows" {
		path = strings.ReplaceAll(path, "/", "\\")
	} else {
		path = strings.ReplaceAll(path, "\\", "/")
	}
	return path
}
