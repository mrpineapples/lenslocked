# !/bin/bash
# Change to the directory with our code that we plan to work from
cd ~/projects/golang/lenslocked

echo "==== Releasing lenslocked ===="
echo "  Deleteing the local binary if it exists (so it isn't uploaded)"
rm lenslocked
echo "  Done!"

echo "  Deleting existing code..."
ssh root@screencast.lens-locked.com "rm -rf /root/go/src/lenslocked"
echo "  Code delete successfully!"

echo "  Uploading code..."
rsync -avr --exclude ".config.json" --exclude '.git' --exclude ".gitignore" --exclude 'tmp' --exclude 'images' ./ root@screencast.lens-locked.com:/root/go/src/lenslocked/
echo "  Code uploaded successfully!"

echo "  Building the code on remote server..."
ssh root@screencast.lens-locked.com "cd /root/app; GO111MODULE=on /usr/local/go/bin/go build -o ./server /root/go/src/lenslocked/*.go"
echo "  Code built successfully!"

echo "  Moving assets..."
ssh root@screencast.lens-locked.com "cd /root/app; cp -R /root/go/src/lenslocked/assets ."
echo "  Assets moved successfully!"

echo "  Moving views..."
ssh root@screencast.lens-locked.com "cd /root/app; cp -R /root/go/src/lenslocked/views ."
echo "  Views moved successfully!"

echo "  Moving Caddyfile..."
ssh root@screencast.lens-locked.com "cd /root/app; cp /root/go/src/lenslocked/Caddyfile ."
echo "  Caddyfile moved successfully!"

echo "  Restarting the server..."
ssh root@screencast.lens-locked.com "sudo service lens-locked.com restart"
echo "  Server restarted successfully!"

echo "  Restarting Caddy server..."
ssh root@screencast.lens-locked.com "sudo service caddy restart"
echo "  Caddy restarted successfully!"

echo "==== Done releasing lenslocked ===="
