#!/bin/bash
Version=1.27.1
go install golang.org/dl/go$Version@latest
go$Version download
