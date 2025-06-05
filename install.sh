#!/bin/bash
mkdir -p /opt/akumanager
mv ./akumanager /opt/akumanager/
mv ./index.html /opt/akumanager/
mv ./static /opt/akumanager/
cp --no-clobber ./akutq_city.conf /etc
cp --no-clobber ./tqstation1.yaml /etc
cp --no-clobber ./akumanager.service /etc/systemd/system/
systemctl enable akumanager.service
systemctl stop akumanager.service
systemctl start akumanager.service
systemctl status akumanager.service
