#!/bin/bash

cmd_probe_av_basic_parameter() {
    file_name=""
    for i in "$@"; do
    case "$i" in
        file=*)
            file_name="${i#*=}"
            shift
            ;;
    esac
    done
    echo "probe filename is $file_name"
    if [ -z "$file_name" ]; then
        echo "$0: file_name is invalid"
        exit 0
    fi
    ffprobe $file_name 2>&1

}

cmd_probe_video_total_frames() {
    file_name=""
    for i in "$@"; do
    case "$i" in
        file=*)
            file_name="${i#*=}"
            shift
            ;;
    esac
    done
    echo "probe filename is $file_name"
    if [ -z "$file_name" ]; then
        echo "$0: file_name is invalid"
        exit 0
    fi
    ffprobe -v error -count_frames -select_streams v:0 -show_entries stream=nb_frames $file_name
}

cmd_probe_audio_total_frames() {
    file_name=""
    for i in "$@"; do
    case "$i" in
        file=*)
            file_name="${i#*=}"
            shift
            ;;
    esac
    done
    echo "probe filename is $file_name"
    if [ -z "$file_name" ]; then
        echo "$0: file_name is invalid"
        exit 0
    fi
    ffprobe -v error -count_frames -select_streams a:0 -show_entries stream=nb_frames $file_name
}


case "$1" in
    cmd=*)
        echo "first parameter is $1"
        op=$1; cmd="${op#*=}"
        shift
        ;;
    help)
        help
        exit 0
        ;;
    *)
        echo "$0 first arg must be cmd=publish/subscribe/report/stop_publish/stop_subscribe"
        exit 0
        ;;
esac

case "$cmd" in
    probe)
        cmd_probe_av_basic_parameter "$@"
        ;;
    *)
        echo "$0 first arg must be cmd=publish/subscribe/report/stop_publish/stop_subscribe"
        exit 0
        ;;
esac