# shipping-labels

A small program written in Go to prepare PDF files from shipping provider into one big file ready to be print.

This program will:

- crop PDF files from Colissimo and Mondial Relay to only keep the shipping label part
- merge everything into one file

## Installation

Grab the corresponding `.tar.gz` from the release page and execute directly the program with

```shell
./shipping-labels
```

## Usage

```shell
Usage:
  shipping-labels [src_dir] [dest_dir] 
```

It will produce a pdf file in `[dest_dir]` named `<YYYYMMDD-hhmm>_ready_to_print.pdf`.