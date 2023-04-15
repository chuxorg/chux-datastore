# Logger Implementation

File Logger is a lightweight asynchronous logging implementatios. It logs messages to files with rotation and compression support. 
The code is designed to be easily extendible for other logging destinations, such as Sentry.io.

## Features

    - Asynchronous logging using goroutines and channels.
    - Log file rotation based on file size.
    - Compressed old log files to save disk space.
    - Easily extendible for other logging destinations.

## Installation

```bash
$ go get -u github.com/yourusername/filelogger
```

