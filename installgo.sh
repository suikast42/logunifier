#!/bin/bash
Version=1.23.6
go install golang.org/dl/go$Version@latest
go$Version download
