package service

import "path/filepath"

const defaultWorkdirName = "grynvault-rebalance"

func defaultWorkdir(dataHome string) string {
	return filepath.Join(dataHome, defaultWorkdirName)
}
