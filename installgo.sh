#!/bin/bash
Version=1.22.3
go install golang.org/dl/go$Version@latest
go$Version download
