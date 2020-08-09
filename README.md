# ffmpego

**ffmpego** is a Go wrapper around the `ffmpeg` command for reading and writing videos. It can be used to programmatically manipulate media with a simple, friendly interface.

# Usage

## Writing a video

To encode a video, create a `VideoWriter` and write `image.Image`s to it. Here's the simplest possible example of encoding a video:

```go
fps := 24.0
width := 50
height := 50

vw, _ := ffmpego.NewVideoWriter("output.mp4", width, height, fps)

for i := 0; i < 24; i++ {
    // Create your image.
    frame := image.NewGray(image.Rect(0, 0, width, height))

    vw.WriteFrame(frame)
}

vw.Close()
```

## Reading a video

Decoding a video is similarly straightforward. Simply create a `VideoReader` and read `image.Image`s from it:

```go
vr, _ := NewVideoReader("input.mp4")

for {
    frame, err := reader.ReadFrame()
    if err == io.EOF {
        break
    }
    // Do something with `frame` here...
}

vr.Close()
```

# Installation

This project depends on the `ffmpeg` command. If you have `ffmpeg` installed, **ffmpego** should already work out of the box.

If you do not already have ffmpeg, you can typically install it using your OS's package manager.

Ubuntu:

```
$ apt install ffmpeg
```

macOS:

```
$ brew install ffmpeg
```

