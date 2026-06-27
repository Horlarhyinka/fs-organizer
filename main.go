package main

import "fmt"

func main() {
	RootCmd.Flags().StringArray("exclude", make([]string, 0), "add files to exclude")
	if err := RootCmd.Execute(); err != nil {
		fmt.Println("error:", err)
	}

}