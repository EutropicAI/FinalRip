package ffmpeg

import (
	"os"
	"os/exec"
	"runtime"

	"github.com/EutropicAI/FinalRip/common/constant"
	"github.com/EutropicAI/FinalRip/module/log"
	"github.com/EutropicAI/FinalRip/module/util"
)

// EncodeVideo 压制视频，压制的 VS 脚本由 encodeScript 指定，压制参数由 encodeParam 指定
func EncodeVideo(encodeScript string, encodeParam string) error {
	encodeScriptPath := "encode.py"

	// 根据操作系统创建脚本文件
	var commandStr string
	var scriptPath string
	var condaInitScript string
	switch runtime.GOOS {
	case OS_WINDOWS:
		scriptPath = "temp_script.bat"
		condaInitScript = "@echo off\r\n" + "call \"%USERPROFILE%\\miniconda3\\condabin\\activate.bat\"\r\n"
	default:
		scriptPath = "temp_script.sh"
		condaInitScript = "#!/bin/bash\n" // Linux 下默认激活 Conda 环境（v0.3后已经移除 Conda 环境
	}
	commandStr = condaInitScript + "vspipe -c y4m " + encodeScriptPath + " - | " + encodeParam
	log.Logger.Info("Encode Command: " + commandStr)

	// 清理临时文件
	_ = util.ClearTempFile(encodeScriptPath, scriptPath)
	defer func(p ...string) {
		log.Logger.Infof("Clear temp file %v", p)
		_ = util.ClearTempFile(p...)
	}(encodeScriptPath, scriptPath)

	// 写入压制 py
	err := os.WriteFile(encodeScriptPath, []byte(encodeScript), 0755)
	if err != nil {
		log.Logger.Error("write vapoursynth script file failed: " + err.Error())
		return err
	}
	// 写入脚本文件
	err = os.WriteFile(scriptPath, []byte(commandStr), 0755)
	if err != nil {
		log.Logger.Error("write script file failed: " + err.Error())
		return err
	}
	// 执行脚本
	var cmd *exec.Cmd
	if runtime.GOOS == OS_WINDOWS {
		cmd = exec.Command("cmd", "/c", scriptPath)
	} else {
		cmd = exec.Command("sh", scriptPath)
	}

	_, err = cmd.CombinedOutput()
	if err != nil {
		log.Logger.Error("vs error: " + err.Error())
		return err
	}
	log.Logger.Info("clip encoded, output: " + constant.FINALRIP_ENCODED_CLIP_MKV)

	return nil
}
