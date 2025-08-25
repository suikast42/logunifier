#!/bin/bash
Version=1.20.0
go install golang.org/dl/go$Version@latest
go$Version download
