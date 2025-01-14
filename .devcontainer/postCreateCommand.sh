#!/usr/bin/env bash

# 将 direnv hook 上 bash
echo 'eval "$(direnv hook bash)"' >> ~/.bashrc
# 将 git 补全加入 .bashrc
echo 'source /usr/share/bash-completion/completions/git' >> ~/.bashrc