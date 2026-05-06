document.addEventListener('DOMContentLoaded', function() {
    // 初始加载服务状态
    // 先加载城市数据，再加载当前城市信息
    fetchCurrentCity();
    fetchCities();
    fetchServicesStatus();
    fetchLEDStatus();
    fetchVolume();
// 页面加载时获取磁盘使用情况
    fetchDiskUsage();
    initRadio();

    // // 每1秒刷新一次服务状态
    setInterval(fetchServicesStatus, 5000);
    setInterval(fetchVolume, 5000);
    setInterval(fetchLEDStatus, 5000);
    setInterval(fetchDiskUsage, 5000);
});


let allStations = [];
let currentIdx = -1;
let currentPlayingName = ""; // 改用名称追踪当前播放
let isPlaying = false;

// 加载电台列表
function loadFMList() {
    return fetch('/api/fm/list')
        .then(res => res.json())
        .then(data => {
            allStations = data.channels;
            renderList(allStations);
            const statusEl = document.getElementById('lcd-status');
            if(statusEl) statusEl.innerText = "UPDATED: " + (data.updatetime || 'NOW');
        });
}

// 确保渲染函数里没有设置特殊的颜色样式
function renderList(stations) {
    const list = document.getElementById('fmList');
    list.innerHTML = '';
    stations.forEach((ch) => {
        const div = document.createElement('div');
        // 根据名称判断高亮，解决搜索后的显示问题
        const isActive = (ch.name === currentPlayingName);
        div.className = `fm-item ${isActive ? 'active' : ''}`;
        div.innerHTML = `<span>📻 ${ch.name}</span>`;

        // 【关键修改点】：绑定点击事件时，直接通过名称查找，不再使用 index
        div.onclick = () => playStationByName(ch.name);
        list.appendChild(div);
    });
}

// 3. 核心修改：通过名称播放
function playStationByName(name) {
    const ch = allStations.find(s => s.name === name);
    if (!ch) return;

    currentPlayingName = name;
    document.getElementById('lcd-name').innerHTML = `<span>${ch.name}</span>`;
    document.getElementById('lcd-status').innerText = "CONNECTING...";

    fetch('/api/fm/play', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({ url: ch.url, name: ch.name })
    }).then(() => {
        isPlaying = true;
        document.getElementById('lcd-status').innerText = "PLAYING";
        document.getElementById('playPauseBtn').innerText = "⏸";
        // 重新渲染列表以更新高亮（无论当前是否在搜索结果中）
        filterStations();
    });
}
// 搜索过滤
function filterStations() {
    const kw = document.getElementById('fmSearch').value.toLowerCase();
    const filtered = allStations.filter(s => s.name.toLowerCase().includes(kw));
    renderList(filtered);
}

// 播放控制
function playStation(index) {
    if (index < 0 || index >= allStations.length) return;
    currentIdx = index;
    const ch = allStations[index];

    document.getElementById('lcd-name').innerHTML = `<span>${ch.name}</span>`;
    document.getElementById('lcd-status').innerText = "CONNECTING...";

    fetch('/api/fm/play', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({ url: ch.url, name: ch.name })
    }).then(() => {
        isPlaying = true;
        document.getElementById('lcd-status').innerText = "PLAYING";
        document.getElementById('playPauseBtn').innerText = "⏸";
        renderList(allStations); // 刷新高亮
    });
}

function togglePlayPause() {
    if (isPlaying) {
        stopFM();
    } else {
        // 逻辑：优先尝试恢复当前记录的名字，如果没有（首次加载），则尝试当前索引
        if (currentPlayingName !== "") {
            playStationByName(currentPlayingName);
        } else if (currentIdx !== -1) {
            playStation(currentIdx);
        } else if (allStations.length > 0) {
            // 如果什么状态都没有，默认播第一个台
            playStationByName(allStations[0].name);
        }
    }
}

function stopFM() {
    fetch('/api/fm/stop', { method: 'POST' }).then(() => {
        isPlaying = false;
        document.getElementById('playPauseBtn').innerText = "▶";
        document.getElementById('lcd-status').innerText = "PAUSED"; // 改为 PAUSED 更像收音机
        // 注意：不要在这里清空 currentPlayingName，否则恢复播放时会找不到目标
    });
}

function nextStation() {
    if (allStations.length === 0) return;
    let idx = allStations.findIndex(s => s.name === currentPlayingName);
    let next = (idx + 1) % allStations.length;
    playStationByName(allStations[next].name);
}

function prevStation() {
    if (allStations.length === 0) return;
    let idx = allStations.findIndex(s => s.name === currentPlayingName);
    let prev = (idx - 1 + allStations.length) % allStations.length;
    playStationByName(allStations[prev].name);
}

// 自动加载
// 修改初始化函数
async function initRadio() {
    await loadFMList(); // 必须 await 确保列表 DOM 生成完毕
    checkCurrentPlaying();
};
function checkCurrentPlaying() {
    fetch('/api/fm/status')
        .then(res => res.json())
        .then(data => {
            if (data.is_playing && data.name) {
                currentPlayingName = data.name;
                document.getElementById('lcd-name').innerHTML = `<span>${data.name}</span>`;
                document.getElementById('lcd-status').innerText = "PLAYING";
                document.getElementById('playPauseBtn').innerText = "⏸";
                isPlaying = true;

                // 刷新高亮显示并滚动到对应位置
                renderList(allStations);
                setTimeout(() => {
                    const activeItem = document.querySelector('.fm-item.active');
                    if (activeItem) activeItem.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
                }, 100);
            }
        });
}
function updateLCDTime() {
    const now = new Date();
    const timeStr = now.getHours().toString().padStart(2, '0') + ":" +
        now.getMinutes().toString().padStart(2, '0');
    const timeEl = document.getElementById('lcd-time');
    if (timeEl) timeEl.innerText = timeStr;
}
setInterval(updateLCDTime, 1000);
function restoreFMStatus() {
    fetch('/api/fm/status')
        .then(res => res.json())
        .then(status => {
            if (status.playing && status.name) {
                // 更新 LCD 显示
                document.getElementById('lcd-name').innerHTML = `<span>${status.name}</span>`;
                document.getElementById('lcd-status').innerText = "PLAYING";
                document.getElementById('playPauseBtn').innerText = "⏸";
                isPlaying = true;

                // 在列表中高亮显示
                const items = document.querySelectorAll('.fm-item');
                items.forEach((item, index) => {
                    if (item.innerText.includes(status.name)) {
                        item.classList.add('active');
                        currentIndex = index; // 恢复当前索引，确保“上下台”功能正常
                        item.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
                    }
                });
            }
        });
}

var cityCode = ""
// 获取服务状态
function fetchServicesStatus() {
    fetch('/api/services')
        .then(response => response.json())
        .then(services => {
            const servicesList = document.getElementById('servicesList');
            servicesList.innerHTML = '';

            services.forEach(service => {
                const serviceItem = document.createElement('div');
                serviceItem.className = 'service-item';

                // 根据服务状态设置不同的样式类
                const statusClass = service.status === 'active' ? 'status-active' :
                    service.status === 'inactive' ? 'status-inactive' : 'status-unknown';

                // 根据服务状态决定是否显示操作按钮
                let actionButtons = '';
                if (service.status === 'unknown') {
                    actionButtons = '';
                } else {
                    actionButtons = service.status === 'active' ?
                        `<button class="stop-btn" onclick="stopService('${service.name}')">停止</button>` :
                        `<button class="start-btn" onclick="startService('${service.name}')">启动</button>`;
                }

                serviceItem.innerHTML = `
                    <div class="service-info">
                        <div class="service-name">${service.name}</div>
                        <div class="service-description">${service.description}</div>
                    </div>
                    <div class="service-actions-container">
                        <span class="service-status ${statusClass}">${service.status}</span>
                        <div class="service-actions">
                            ${actionButtons}
                        </div>
                    </div>
                `;

                servicesList.appendChild(serviceItem);
            });
        })
        .catch(error => console.error('获取服务状态失败:', error));
}

// 启动服务
function startService(name) {
    fetch(`/api/service/${name}/start`, {
        method: 'POST'
    })
        .then(response => response.json())
        .then(data => {
            console.log(data.message);
            fetchServicesStatus();
        })
        .catch(error => console.error('启动服务失败:', error));
}

// 停止服务
function stopService(name) {
    fetch(`/api/service/${name}/stop`, {
        method: 'POST'
    })
        .then(response => response.json())
        .then(data => {
            console.log(data.message);
            fetchServicesStatus();
        })
        .catch(error => console.error('停止服务失败:', error));
}

// 获取音量
function fetchVolume() {
    fetch('/api/volume')
        .then(response => response.json())
        .then(volume => {
            document.getElementById('volumeValue').textContent = volume;
            document.getElementById('volumeSlider').value = volume;
        })
        .catch(error => console.error('获取音量失败:', error));
}

// 调整音量
function adjustVolume(value) {
    // 确保音量在1-63的范围内
    const volume = Math.min(Math.max(value, 1), 63);

    fetch('/api/volume', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ volume: volume })
    })
        .then(response => response.json())
        .then(data => {
            document.getElementById('volumeValue').textContent = volume;
        })
        .catch(error => console.error('设置音量失败:', error));
}

// 获取LED灯状态
function fetchLEDStatus() {
    fetch('/api/led')
        .then(response => response.json())
        .then(ledStatus => {
            const ledSwitch = document.getElementById('ledSwitch');
            if (ledStatus) {
                ledSwitch.classList.add('active');
            } else {
                ledSwitch.classList.remove('active');
            }
        })
        .catch(error => console.error('获取LED灯状态失败:', error));
}

// 切换LED灯状态
function toggleLED() {
    const ledSwitch = document.getElementById('ledSwitch');
    const isActive = ledSwitch.classList.contains('active');

    fetch('/api/led', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ state: !isActive })
    })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                if (data.state) {
                    ledSwitch.classList.add('active');
                } else {
                    ledSwitch.classList.remove('active');
                }
            }
        })
        .catch(error => console.error('切换LED灯状态失败:', error));
}

// 获取所有城市
function fetchCities() {
    fetch('/api/cities')
        .then(response => response.json())
        .then(cities => {
            const citySelect = document.getElementById('citySelect');
            citySelect.innerHTML = ''; // 清空选项

            cities.forEach(city => {
                const option = document.createElement('option');
                option.value = city.Code;
                option.textContent = `${city.City} (${city.Province})`;
                if (city.Code === cityCode) {
                    option.selected = true;
                }
                citySelect.appendChild(option);
            });
        })
        .catch(error => console.error('获取城市数据失败:', error));
}

// 获取当前城市信息
function fetchCurrentCity() {
    fetch('/api/currentcity')
        .then(response => response.json())
        .then(cityData => {
            const cityInfo = document.getElementById('cityInfo');
            cityInfo.innerHTML = `
                <div>当前城市: ${cityData.city}</div>
                <div>城市代码: ${cityData.code}</div>
                <div>省份: ${cityData.province}</div>
            `;

            // 设置默认选中项
            const citySelect = document.getElementById('citySelect');
            if (citySelect) {
                citySelect.value = cityData.code;
                cityCode = cityData.code;
            }
        })
        .catch(error => {
            console.error('获取当前城市失败:', error);
            return Promise.reject(error);
        });
}

// 更新城市代码
function updateCityCode() {
    const citySelect = document.getElementById('citySelect');
    const selectedCode = citySelect.value;

    fetch('/api/city', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ code: selectedCode })
    })
        .then(response => {
            if (!response.ok) {
                throw new Error('更新城市失败');
            }
            return response.json();
        })
        .then(data => {
            // 显示成功消息
            Swal.fire({
                icon: 'success',
                title: '成功',
                text: data.message
            });
            fetchCurrentCity(); // 更新当前城市显示
        })
        .catch(error => {
            // 显示失败消息
            Swal.fire({
                icon: 'error',
                title: '错误',
                text: '更新城市失败: ' + error.message
            });
            console.error('更新城市失败:', error);
        });
}

// 获取磁盘使用情况
function fetchDiskUsage() {
    fetch('/api/diskusage')
        .then(response => response.json())
        .then(data => {
            const diskUsageText = document.querySelector('.disk-usage-text');
            const diskUsageProgress = document.querySelector('.disk-usage-progress');

            diskUsageText.textContent = `已用: ${data.used} GB / 总计: ${data.total} GB`;
            diskUsageProgress.style.width = `${data.percentage}%`;
        })
        .catch(error => {
            console.error('获取磁盘使用情况失败:', error);
            const diskUsageText = document.querySelector('.disk-usage-text');
            diskUsageText.textContent = '获取磁盘使用情况失败';
        });
}