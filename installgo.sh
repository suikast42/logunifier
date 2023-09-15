#!/bin/bash
Version=1.21.1
go install golang.org/dl/go$Version@latest
go$Version download
