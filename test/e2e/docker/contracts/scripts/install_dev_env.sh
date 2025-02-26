#!/bin/bash


apt-get install -yqq iputils-ping lsof net-tools vim-common tree expect

npm install --global hardhat-shorthand

echo "alias ls='ls --color' " >> ~/.bashrc && echo "alias ll='ls -l' " >> ~/.bashrc
echo '#!/usr/bin/expect' > /tmp/install_hardhat_completion.exp && \
  echo 'set time 10' >> /tmp/install_hardhat_completion.exp && \
  echo 'spawn hardhat-completion install' >> /tmp/install_hardhat_completion.exp && \
  echo 'expect "Which Shell do you use?"' >> /tmp/install_hardhat_completion.exp && \
  echo 'send -- "bash\r"' >> /tmp/install_hardhat_completion.exp && \
  echo 'expect "We will install completion to ~/.bashrc, is it ok? (y/N)"' >> /tmp/install_hardhat_completion.exp && \
  echo 'send -- "y\r"' >> /tmp/install_hardhat_completion.exp && \
  echo 'expect eof' >> /tmp/install_hardhat_completion.exp && \
  /usr/bin/expect /tmp/install_hardhat_completion.exp