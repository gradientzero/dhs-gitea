#!/bin/bash

# Copy files from old server
scp -r root@95.217.101.177:/var/lib/docker/volumes/usb_db-data/ /home/xubuntu/usb_db-data
scp -r root@95.217.101.177:/home/user/git/ /home/xubuntu/git
scp -r root@95.217.101.177:/home/user/avatars/ /home/xubuntu/avatars

# Paste files to new server
scp -r /home/xubuntu/usb_db-data root@157.90.23.177:/home/user/dhcs/usb_db-data
scp -r /home/xubuntu/git root@157.90.23.177:/home/user/dhcs/git
scp -r /home/xubuntu/avatars root@157.90.23.177:/home/user/dhcs/avatars
