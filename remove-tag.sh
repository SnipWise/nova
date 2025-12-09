#!/bin/bash
set -o allexport; source release.env; set +o allexport

git tag -d ${TAG}


