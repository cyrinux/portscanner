#!/bin/sh
tor &
grpcnmapscanner "$@"
