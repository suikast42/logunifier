#!/bin/bash
Version=1.20.2
go install golang.org/dl/go$Version@latest
go$Version download
