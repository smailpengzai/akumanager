document.addEventListener('DOMContentLoaded', function() {
    // 初始加载服务状态
    fetchServicesStatus();

    // 每30秒刷新一次服务状态
    setInterval(fetchServicesStatus, 30000);

    // 初始化音量控制
    fetchVolume();
});

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

// 在页面加载时初始化音量
fetchVolume();

document.addEventListener('DOMContentLoaded', function() {
    // 初始加载服务状态
    fetchServicesStatus();

    // 每30秒刷新一次服务状态
    setInterval(fetchServicesStatus, 30000);

    // 初始化音量控制
    fetchVolume();

    // 初始化LED灯状态
    fetchLEDStatus();
});

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