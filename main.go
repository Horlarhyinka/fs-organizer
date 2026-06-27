package main

import "fmt"

func main() {
	RootCmd.Flags().StringArray("exclude", make([]string, 0), "add files to exclude")
	RootCmd.Flags().String("out", "out", "output directory")
	if err := RootCmd.Execute(); err != nil {
		fmt.Println("error:", err)
	}

}