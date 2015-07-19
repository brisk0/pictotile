pictotile
=========

Simple tool for converting images to gameboy tile data.

Installation
------------

Ensure the golang toolchain is installed on your machine, then run from the repository root:

> $ go build pictotile.go

This produces a self-contained binary and should be able to be placed and run anywhere.

Note that there are binaries in the Snapshots directory. These are outdated due to problems with the cross compiler. The binary in the root directory is current and built for linux, amd_64. Building for yourself is always preferable, and should always work.

Usage
-----

	pictotile [options] [infile [outfile]]

Inputs
------

Pictotile accepts png, jpeg and gif (not animated) file formats.

Input and output default to stdin/stdout, which can be set explicitly using "-" for either argument, otherwise the program can both read from and write to files.

Options
-------

-d, --dim=1: Square dimension in number tiles of each sprite

-w, --width=1: Width of each sprite in number of tiles

-h, --height=1: Height of each tile in number of tiles

-o, --offset=0: Offset of the first tile from both the top and left edge

-x, --xoffset=0: Horizontal offset of first tile from left

-y, --yoffset=0: Vertical offset of first tile from top

-s, --spacing=0: Distance between sprites

-X, --xspacing=0: Horizontal distance between sprites

-Y, --yspacing=0: Vertical distance between sprites

-f, --format="0x%X, ": C Style format for output data (printed in a loop for each byte)

-t, --spritemode=false: Sets first color in tile as transparency (color 0)

Description
-----------

Pictotile converts image files into gameboy compatible format. Default format prints each byte in the format "0xFF, ", but format can be specified using -f.

Tile data is output in a left-right, top-bottom order, however, tiles within a sprite are all output one after another.

For example, the command

	pictotile -d2 img.png img.tile

will put the following set of tiles:

	img.png:
	+-+-+-+-+
	|0|1|2|3|
	+-+-+-+-+
	|4|5|6|7|
	+-+-+-+-+
	
in the following order:

	img.tile:
	0145 2367
