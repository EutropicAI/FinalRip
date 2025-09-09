package ffmpeg

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"

	"github.com/EutropicAI/FinalRip/module/log"
	"github.com/EutropicAI/FinalRip/module/util"
)

type FileInfo struct {
	Path  string
	Index int
}

// CutVideo 使用 ffmpeg 进行视频切割，每段视频时长为 60s
func CutVideo(inputPath string, outputFolder string) ([]string, error) {
	// 根据操作系统创建脚本文件
	var commandStr string
	var scriptPath string
	switch runtime.GOOS {
	case OS_WINDOWS:
		commandStr = fmt.Sprintf("ffmpeg -i \"%s\" -f segment -segment_format mkv -segment_time 60 -reset_timestamps 1 -c copy -map 0:v:0 -segment_list \"%s/out.list\" \"%s/%%%%003d.mkv\"", inputPath, outputFolder, outputFolder) //nolint: lll
		scriptPath = "temp_script.bat"
		commandStr = fmt.Sprintf("@echo off%s%s", "\r\n", commandStr)
	default:
		commandStr = fmt.Sprintf("ffmpeg -i \"%s\" -f segment -segment_format mkv -segment_time 60 -reset_timestamps 1 -c copy -map 0:v:0 -segment_list \"%s/out.list\" \"%s/%%003d.mkv\"", inputPath, outputFolder, outputFolder) //nolint: lll
		scriptPath = "temp_script.sh"
		commandStr = fmt.Sprintf("#!/bin/bash%s%s", "\n", commandStr)
	}

	// 清理临时文件
	_ = util.ClearTempFile(scriptPath)
	defer func(p ...string) {
		log.Logger.Infof("Clear temp file %v", p)
		_ = util.ClearTempFile(p...)
	}(scriptPath)

	// 写入脚本文件
	err := os.WriteFile(scriptPath, []byte(commandStr), 0755)
	if err != nil {
		log.Logger.Errorf("write script file failed: %v", err)
		return nil, err
	}

	// 执行脚本
	var cmd *exec.Cmd
	if runtime.GOOS == OS_WINDOWS {
		cmd = exec.Command("cmd", "/c", scriptPath)
	} else {
		cmd = exec.Command("sh", scriptPath)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Logger.Info(err.Error())
	}
	log.Logger.Info(string(out))

	var outputFiles []FileInfo
	// 遍历输出文件列表，读取文件夹下的所有文件
	err = filepath.Walk(outputFolder, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && filepath.Ext(path) == ".mkv" {
			base := filepath.Base(path)
			ext := filepath.Ext(base)
			// 从文件名中提取 index，去掉后缀
			index, err := strconv.Atoi(base[:len(base)-len(ext)])
			if err != nil {
				log.Logger.Errorf("Failed to parse index from file %s: %v", path, err)
				return nil
			}
			outputFiles = append(outputFiles, FileInfo{
				Path:  path,
				Index: index,
			})
		}
		return nil
	})
	if err != nil {
		log.Logger.Errorf("Failed to walk output folder: %v", err)
		return nil, err
	}

	sort.Slice(outputFiles, func(i, j int) bool {
		return outputFiles[i].Index < outputFiles[j].Index
	})

	var outputPaths []string
	for _, file := range outputFiles {
		outputPaths = append(outputPaths, file.Path)
	}

	return outputPaths, nil
}
