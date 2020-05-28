package main

import (
    "fmt"
    "math"
    "os"
    "os/exec"
    "strconv"
    "strings"
    "sync"
)
// 判断所给路径文件/文件夹是否存在
func Exists(path string) bool {
    _, err := os.Stat(path)    //os.Stat获取文件信息
    if err != nil {
        if os.IsExist(err) {
            return true
        }
        return false
    }
    return true
}

// 判断所给路径是否为文件夹
func IsDir(path string) bool {
    s, err := os.Stat(path)
    if err != nil {
        return false
    }
    return s.IsDir()
}

// 判断所给路径是否为文件
func IsFile(path string) bool {
    return !IsDir(path)
}

func main(){
    if len(os.Args) != 2 {
        fmt.Println("输入一个参数作为视频文件, 使用方法: ./av filename")
        return
    }
    filename := os.Args[1]
    if !Exists(filename) {
        fmt.Println("文件不存在")
        return
    }
    if !IsFile(filename) {
        fmt.Println("没有指定一个文件")
        return
    }
    exsit_audio := false
    exsit_video := false
    actual_audio_frames := 0
    actual_video_frames := 0
    duration := 0.0
    video_fps := 0.0
    audio_sample_rate:=0.0

    body, err := exec.Command("ffprobe", filename).CombinedOutput()
    //rs, err := sh.Command("/Users/dushuanglong/go/src/learn/av/probe_av.sh","cmd=probe", fmt.Sprintf("file=%s", filename)).Output()
    if err != nil {
        fmt.Printf("获取视频Metadata信息失败, err: %s\n", err)
        return
    }
    result := string(body)
    //fmt.Printf("%s\n", result)
    lines := strings.Split(result, "\n")
    for _, line := range lines {
        //Duration: 00:00:39.42, start: 0.000000, bitrate: 612 kb/s
        if strings.Contains(line, "Duration:") {
            //fmt.Printf("***Duration: %s\n", line)
            parts := strings.Split(line, ",")
            for _, part := range parts {
                //Duration: 00:00:39.42
                if strings.Contains(part, "Duration") {
                    part = strings.TrimSpace(part)
                    times := strings.Split(part, " ")[1]
                    //00:00:39.42
                    for _, time := range strings.Split(times, ":") {
                        t, _ := strconv.ParseFloat(time, 32)
                        duration = duration*60 + t
                    }
                }
            }
        }
        //Stream #0:0(und): Video: h264 (High) (avc1 / 0x31637661), yuv420p, 640x480, 477 kb/s, 24.91 fps, 25 tbr, 16k tbn, 50 tbc (default)
        if strings.Contains(line, "Video:") {
            exsit_video = true
            //fmt.Printf("***Video: %s\n", line)
            parts := strings.Split(line, ",")
            for _, part := range parts {
                //24.91 fps
                if strings.Contains(part, "fps") {
                    part = strings.TrimSpace(part)
                    FPS := strings.Split(part, " ")[0]
                    video_fps, _ = strconv.ParseFloat(FPS, 32)
                }
            }
        }
        //Stream #0:1(und): Audio: aac (LC) (mp4a / 0x6134706D), 44100 Hz, stereo, fltp, 128 kb/s (default)
        if strings.Contains(line, "Audio:") {
            exsit_audio = true
            //fmt.Printf("***Audio: %s\n", line)
            parts := strings.Split(line, ",")
            for _, part := range parts {
                //44100 Hz
                if strings.Contains(part, "Hz") {
                    part = strings.TrimSpace(part)
                    audio_sample_rate, _ = strconv.ParseFloat(strings.Split(part, " ")[0], 32)
                }
            }
        }
    }
    //fmt.Printf("duration: %.2f, fps: %.2f, sample_rate: %.2f\n", duration, video_fps, audio_sample_rate)
    if !exsit_audio || !exsit_video {
        if exsit_video {
            fmt.Println("只存在视频")
        }
        if exsit_audio {
            fmt.Println("只存在音频")
        }
        return
    }
    err = nil
    var Wg sync.WaitGroup
    Wg.Add(2)
    //视频总帧数
    go func() {
        defer Wg.Done()
        rs, err  := exec.Command("ffprobe", "-v", "error", "-count_frames", "-select_streams", "v:0", "-show_entries", "stream=nb_frames", filename).Output()
        if err != nil {
            fmt.Printf("获取视频帧数信息失败, err: %s\n", err)
        }
        result := string(rs)
        //fmt.Printf("audio frames******%s\n", result)
        lines := strings.Split(result, "\n")
        for _, line := range lines {
            if strings.Contains(line, "nb_frames") {
                actual_video_frames, _ = strconv.Atoi(strings.Split(line, "=")[1])
            }
        }
    }()
    //音频总帧数
    go func() {
        defer Wg.Done()
        rs, err  := exec.Command("ffprobe", "-v", "error", "-count_frames", "-select_streams", "a:0", "-show_entries", "stream=nb_frames", filename).Output()
        if err != nil {
            fmt.Printf("获取音频帧数信息失败, err: %s\n", err)
        }
        result := string(rs)
        //fmt.Printf("video frames******%s\n", result)
        lines := strings.Split(result, "\n")
        for _, line := range lines {
            if strings.Contains(line, "nb_frames") {
                actual_audio_frames, _ = strconv.Atoi(strings.Split(line, "=")[1])
            }
        }
    }()
    Wg.Wait()
    if err != nil {
        fmt.Println("解析音视频帧数信息时失败")
        return
    }
    count_video_frames := 0
    count_audio_frames := 0
    count_video_frames = int(video_fps * duration)
    count_audio_frames = int(audio_sample_rate * duration) / 1024
    //fmt.Printf("count_video_frames: %d\n", count_video_frames)
    //fmt.Printf("count_audio_frames: %d\n", count_audio_frames)
    //标记符号，理论帧数 >= 实际帧数时，置为true
    video_frame_diff_flag := false
    audio_frame_diff_flag := false
    video_frame_diff := 0
    audio_frame_diff := 0
    //精确到ms
    video_diff_time := 0.0
    audio_diff_time := 0.0
    if count_video_frames >= actual_video_frames {
        video_frame_diff_flag = true
        video_frame_diff = count_video_frames - actual_video_frames
    } else {
        video_frame_diff_flag = false
        video_frame_diff = actual_video_frames - count_video_frames
    }
    video_diff_time = float64(video_frame_diff * 100) / video_fps
    if count_audio_frames >= actual_audio_frames {
        audio_frame_diff_flag = true
        audio_frame_diff = count_audio_frames - actual_audio_frames
    } else {
        audio_frame_diff_flag = false
        audio_frame_diff = actual_audio_frames - count_audio_frames
    }
    audio_diff_time = float64(audio_frame_diff * 1024) / audio_sample_rate
    final_diff := 0
    //计算最后的音视频偏差
    //符号相反，累加
    if video_frame_diff_flag != audio_frame_diff_flag {
        final_diff = int(video_diff_time + audio_diff_time)
    } else {
        //符号相同，相减
        float_diff := math.Abs(video_diff_time - audio_diff_time)
        final_diff = int(float_diff)
    }
    result = fmt.Sprintf("视频总时长: %.2f s \n理论视频帧数: %d, 理论音频帧数: %d \n实际视频帧数: %d, 实际音频帧数: %d \n音视频总体不同步时间为: %d ms\n", duration, count_video_frames, count_audio_frames, actual_video_frames, actual_audio_frames, final_diff)
    fmt.Print(result)
  }

