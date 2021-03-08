# #!/bin/bash
set -ue -o pipefail
export LC_ALL=C

cd ${HOME}

echo "Start bootstrap..."
echo ""

####################################################################################################
# Install utility packages
####################################################################################################

# via package manager
sudo apt-get -qq update && sudo apt-get -qq install -y \
  git \
  wget \
  vim \
  curl \
  dnsutils \
  net-tools \
  bash-completion \
  silversearcher-ag \
  unzip \
  nginx \
  dstat \
  tcpdump \
  htop \
  fio

# kataribe
go get -u github.com/matsuu/kataribe
cat << 'EOF' >> ${HOME}/.bashrc

# set PATH for kataribe
export GOPATH="${HOME}/go"
export PATH="${PATH}:${GOPATH}/bin"
EOF

echo "Utility packages are installed."
echo ""

####################################################################################################
# Prompt settings
####################################################################################################

cat << 'EOF' >> ~/.bashrc

# modify prompt
if type __git_ps1 > /dev/null 2>&1 ; then
  export GIT_PS1_SHOWDIRTYSTATE=true
  export GIT_PS1_SHOWSTASHSTATE=true
  export GIT_PS1_SHOWUNTRACKEDFILES=true
  export GIT_PS1_SHOWUPSTREAM="auto"
  export GIT_PS1_SHOWCOLORHINTS=true
fi
export PS1='[\e[01;32m\u@\h\e[0;00m:\w\e[01;31m$(__git_ps1)\e[0;00m \t EXIT=$?] \n\$ '
EOF

echo "A prompt setting is updated. To apply change, reload a terminal."
echo ""

####################################################################################################
# Environment settings
####################################################################################################

# add a user to adm group
sudo usermod -aG adm `id -un`

echo "To read protected logs, a current user is added to adm group."
echo ""

# set Git config
git config --global user.name "`hostname`"
git config --global user.email "b4train@example.com"

echo "Git config is set as follows:"
git config --global user.name
git config --global user.email
echo ""

# create private/public keys for GitHub
ssh-keygen -q -t ed25519 -f "${HOME}/.ssh/git_ed25519" -N ""
cat << EOF >> ${HOME}/.ssh/config

Host github.com
  User git
  IdentityFile ${HOME}/.ssh/git_ed25519
EOF

echo "A new SSH-key is created for GitHub. Register the following public key."
cat ${HOME}/.ssh/git_ed25519.pub
echo ""

echo "Done."
