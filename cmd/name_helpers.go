package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func loadAgentNames(flagNames []string, namesFile string) ([]string, error) {
	names := make([]string, 0, len(flagNames))
	for _, name := range flagNames {
		name = strings.TrimSpace(name)
		if name != "" {
			names = append(names, name)
		}
	}

	if strings.TrimSpace(namesFile) == "" {
		return names, nil
	}

	path, err := resolvePath(namesFile)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		names = append(names, line)
	}
	return names, nil
}

func collectAgentNames(prompts []string) ([]string, error) {
	reader := bufio.NewReader(os.Stdin)
	names := make([]string, 0, len(prompts))
	for _, prompt := range prompts {
		if strings.TrimSpace(prompt) == "" {
			continue
		}
		fmt.Printf("%s: ", strings.TrimSpace(prompt))
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		for _, part := range strings.Split(strings.TrimSpace(line), ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				names = append(names, part)
			}
		}
	}
	return names, nil
}
