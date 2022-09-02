<img src="logo.jpg" width="160">

WIP PSX emulator in Go

![](startup.gif)

![](psxniccc.gif)

TODO:
- Figure out how to load game rom
- CDROM
- timers
- CPU
  * pass amidog's psx_cpu test
  * fix issues with load delay slots??
  * GTE coprocessor
- timing
  * instruction caching
  * Using [this](https://github.com/JaCzekanski/ps1-tests/blob/master/cpu/access-time/psx.log) as reference, implement bus waitstates
- better DMA behaviour
- controller input
- more GPU commands and other obscure GPU stuff
- optimize software renderer
  * use fixed point arithmetic
  * run renderer in a different thread?
  * many other things
- boot Crash Bandicoot
- web server for debugging
- wasm port
- savestates
