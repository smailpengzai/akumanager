#!/bin/bash
mkdir -p /opt/akumanager
mv -f ./akumanager /opt/akumanager/
mv -f ./index.html /opt/akumanager/
rm -rf /opt/akumanager/static
mv -f ./static /opt/akumanager/
cp --no-clobber ./akutq_city.conf /etc
cp --no-clobber ./tqstation1.yaml /etc
cp --no-clobber ./akumanager.service /etc/systemd/system/
systemctl enable akumanager.service
systemctl stop akumanager.service
systemctl start akumanager.service
systemctl status akumanager.service
