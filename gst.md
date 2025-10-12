
v4l2-ctl --list-devices

v4l2-ctl --device=/dev/video0 --all


```bash
#mppjpegdec

gst-launch-1.0 v4l2src device=/dev/video0 ! \
'image/jpeg, width=1920, height=1080, framerate=30/1' ! \
jpegparse ! \
mppjpegdec ! video/x-raw, format=NV12 ! videoconvert ! fakesink

GST_DEBUG=fpsdisplaysink:5 gst-launch-1.0 v4l2src device=/dev/video0 ! \
'image/jpeg, width=1920, height=1080, framerate=30/1' ! \
jpegparse ! \
mppjpegdec ! video/x-raw, format=NV12 ! \
videoconvert ! \
fpsdisplaysink text-overlay=false video-sink=fakesink sync=false

gst-launch-1.0 v4l2src device=/dev/video0 ! \
'image/jpeg, width=1920, height=1080, framerate=30/1' ! \
jpegparse ! \
mppjpegdec ! video/x-raw, format=NV12 ! \
mpph265enc ! h265parse ! mp4mux ! filesink location=output.mp4
```