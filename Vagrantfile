# -*- mode: ruby -*-
# vi: set ft=ruby :

# Vagrantfile API/syntax version. Don't touch unless you know what you're doing!
VAGRANTFILE_API_VERSION = "2"

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
  config.vm.box = "opscode-ubuntu-13.04"
  config.vm.box_url = "https://opscode-vm-bento.s3.amazonaws.com/vagrant/opscode_ubuntu-13.04_provisionerless.box"
  config.vm.network :forwarded_port, guest: 4243, host: 14243
  config.vm.provision :shell, inline: <<'SHELL'
wget -O - http://get.docker.io/gpg | apt-key add -
echo deb http://get.docker.io/ubuntu docker main > /etc/apt/sources.list.d/docker.list
apt-get update
apt-get install --yes linux-image-extra-`uname -r` lxc-docker
sed -i 's/\/usr\/bin\/docker -d/\/usr\/bin\/docker -H=0.0.0.0:4243 -d/' /etc/init/docker.conf
service docker restart
sleep 5
docker -H=127.0.0.1:4243 pull ubuntu:precise
SHELL
end
