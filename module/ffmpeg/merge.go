package ffmpeg

import (
	"fmt"
	"os/exec"

	"github.com/EutropicAI/FinalRip/module/log"
	"github.com/EutropicAI/FinalRip/module/util"
)

// MergeVideo 使用 ffmpeg 进行视频合并，使用 mkvpropedit 清除 tags
func MergeVideo(originPath string, inputFiles []string, outputPath string) error {
	tempVideoConcatOutputPath := "temp_video_concat_output.mkv"
	tempVideoMergedOutputPath := "temp_video_merged_output.mkv"

	// clear temp file
	_ = util.ClearTempFile(tempVideoConcatOutputPath, tempVideoMergedOutputPath)
	defer func(p ...string) {
		log.Logger.Infof("Clear temp file %v", p)
		_ = util.ClearTempFile(p...)
	}(tempVideoConcatOutputPath, tempVideoMergedOutputPath)

	// Concat video
	log.Logger.Infof("Concat video with encoded clips: %s", inputFiles)

	mkvmergeArgs := []string{"-v", "-o", tempVideoConcatOutputPath}
	for i, file := range inputFiles {
		if i > 0 {
			mkvmergeArgs = append(mkvmergeArgs, "+")
		}
		mkvmergeArgs = append(mkvmergeArgs, file)
	}

	cmd := exec.Command("mkvmerge", mkvmergeArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Logger.Errorf("Concat video failed: %v", err)
		return err
	}
	log.Logger.Infof("Concat video output: %s", out)

	// ReMux with source video
	err = ReMuxWithSourceVideo(originPath, tempVideoMergedOutputPath, tempVideoConcatOutputPath)
	if err != nil {
		log.Logger.Errorf("ReMux with source video failed: %v", err)
		return err
	}

	// ReMux with mkvmerge
	// !mkvmerge -o output.mkv temp_merged.mkv
	log.Logger.Infof("ReMux video with mkvmerge...")
	cmd = exec.Command(
		"mkvmerge",
		"-o", outputPath,
		tempVideoMergedOutputPath,
	)
	out, err = cmd.CombinedOutput()
	if err != nil {
		log.Logger.Errorf("Re-mux video failed: %v", err)
		return err
	}
	log.Logger.Infof("Re-mux output: %s", out)

	// Remove tags with mkvpropedit
	// !mkvpropedit output.mkv --tags all:
	log.Logger.Infof("Remove tags with mkvpropedit...")
	cmd = exec.Command(
		"mkvpropedit",
		outputPath,
		"--tags", "all:",
	)
	out, err = cmd.CombinedOutput()
	if err != nil {
		log.Logger.Errorf("Remove tags failed: %v", err)
		return err
	}
	log.Logger.Infof("Remove tags output: %s", out)
	return nil
}

// ReMuxWithSourceVideo 使用 ffmpeg 和原始视频进行 remux
func ReMuxWithSourceVideo(originPath string, tempVideoMergedOutputPath string, concatOutputPath string) error {
	// Define the different codec combinations to try
	codecCombinations := [][]string{
		// Try copying both audio and subtitles
		{
			"-map", "1:v:0",
			"-map", "0:a",
			"-map", "0:s",
			"-disposition:v:0", "default",
			"-c:v", "copy",
			"-c:a", "copy",
			"-c:s", "copy",
			"-max_interleave_delta", "0",
		},
		// Try FLAC for audio and copy subtitles
		{
			"-map", "1:v:0",
			"-map", "0:a",
			"-map", "0:s",
			"-disposition:v:0", "default",
			"-c:v", "copy",
			"-c:a", "flac",
			"-c:s", "copy",
			"-max_interleave_delta", "0",
		},
		// Try copying only audio
		{
			"-map", "1:v:0",
			"-map", "0:a",
			"-disposition:v:0", "default",
			"-c:v", "copy",
			"-c:a", "copy",
			"-max_interleave_delta", "0",
		},
		// Try FLAC for audio only
		{
			"-map", "1:v:0",
			"-map", "0:a",
			"-disposition:v:0", "default",
			"-c:v", "copy",
			"-c:a", "flac",
			"-max_interleave_delta", "0",
		},
	}

	// Iterate over each codec combination
	for _, codecArgs := range codecCombinations {
		// Build the ffmpeg command with the current codec combination
		args := []string{
			"-i", originPath,
			"-i", concatOutputPath,
		}
		args = append(args, codecArgs...)
		args = append(args, tempVideoMergedOutputPath)

		cmd := exec.Command("ffmpeg", args...)
		out, err := cmd.CombinedOutput()
		log.Logger.Infof("ffmpeg remux output with codec combination %v: %s", codecArgs, out)

		if err == nil {
			// Success, return nil
			return nil
		}

		// Log the error and try the next combination
		log.Logger.Errorf("ffmpeg remux failed with codec combination %v: %v", codecArgs, err)
		// If failed, clean up temp files which may have been created
		_ = util.ClearTempFile(tempVideoMergedOutputPath)
	}

	return fmt.Errorf("ffmpeg remux: all codec combinations failed")
}
