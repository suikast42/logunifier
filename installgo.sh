#!/bin/bash
Version=1.21.5
go install golang.org/dl/go$Version@latest
go$Version download
