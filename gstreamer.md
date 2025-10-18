Maybe useful? https://forum.armbian.com/topic/52513-manually-adding-new-mesa-mali-drivers-to-armbian-debain-612/

```bash
apt update 
apt install -y git make cmake gcc g++ wget python3-pip pipx ninja-build
pipx install meson

export PATH=$PATH:/root/.local/bin

#Install GST
apt install -y gstreamer1.0-plugins-bad gstreamer1.0-plugins-base gstreamer1.0-tools gstreamer1.0-plugins-good libgstreamer-plugins-bad1.0-dev libgstreamer-plugins-base1.0-dev libgstreamer1.0-dev


#Build MPP
git clone https://github.com/rockchip-linux/mpp -b develop
cd ~/mpp/build/linux/aarch64
git checkout 1.0.11
./make-Makefiles.bash
make -j4
make install

#Build librga-rockchip
git clone https://github.com/tsukumijima/librga-rockchip.git
meson setup librga-rockchip rkrga_build \
    --prefix=/usr \
    --libdir=lib \
    --buildtype=release \
    --default-library=shared \
    -Dcpp_args=-fpermissive \
    -Dlibdrm=false \
    -Dlibrga_demo=false
meson configure rkrga_build
ninja -C rkrga_build install

#Istall gstreamer-rockchip
git clone -b gstreamer-rockchip https://github.com/JeffyCN/mirrors.git --depth=1
cd mirrors
mkdir build
meson build
ninja -C build install

#COPY gstreamer-rockchip to correct directory
cp -r /usr/local/lib/aarch64-linux-gnu/* /usr/lib/aarch64-linux-gnu/
cp -r /usr/local/lib/gstreamer-1.0/* /usr/lib/aarch64-linux-gnu/gstreamer-1.0


#Verify gstreamer-rockchip
gst-inspect-1.0 | grep mpp

#
sudo tee /etc/udev/rules.d/99-rockchip-mpp.rules > /dev/null <<'EOF'
# rockchip mpp device perms:
KERNEL=="mpp_service", GROUP="video", MODE="0660"

# apply to the dma_heap parent device node and per-heap children
SUBSYSTEM=="dma_heap", KERNEL=="dma_heap", GROUP="video", MODE="0660"
SUBSYSTEM=="dma_heap", KERNEL=="system-uncached", GROUP="video", MODE="0660"
SUBSYSTEM=="dma_heap", KERNEL=="system", GROUP="video", MODE="0660"



# fallback: match any dma_heap child (if kernel names vary)
SUBSYSTEM=="dma_heap", RUN+="/bin/chgrp video /dev/dma_heap/%k", RUN+="/bin/chmod 0660 /dev/dma_heap/%k"
EOF

sudo udevadm control --reload-rules
sudo udevadm trigger

sudo usermod -a -G video $(whoami)
sudo reboot
```