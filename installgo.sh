#!/bin/bash
Version=1.24.5
go install golang.org/dl/go$Version@latest
go$Version download
