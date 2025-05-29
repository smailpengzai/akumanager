package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

type ServiceStatus struct {
	Name        string `json:"name"`
	Status      string `json:"status"`
	Description string `json:"description"`
}

var (
	services = []ServiceStatus{
		{Name: "sysboot", Description: "系统启动服务"},
		{Name: "showtime", Description: "时间显示服务"},
		{Name: "akuweb", Description: "网页控制台服务"},
		{Name: "akuapp", Description: "APP后台服务"},
		{Name: "fm", Description: "网络收音机服务"},
		{Name: "ttyd", Description: "网页终端服务"},
		{Name: "shairport-sync", Description: "AirPlay服务"},
		{Name: "xiaozhi", Description: "小智语音服务"},
		{Name: "dlna", Description: "DLNA服务"},
		{Name: "alist", Description: "文件服务"},
	}
	serviceStatusMap = make(map[string]ServiceStatus)
	muxLock          sync.Mutex
)

func main() {
	// 初始化服务状态
	for _, service := range services {
		serviceStatusMap[service.Name] = service
	}

	// 启动服务状态监控
	go monitorServices()

	r := mux.NewRouter()

	// API路由
	r.HandleFunc("/api/services", GetServicesHandler).Methods("GET")
	r.HandleFunc("/api/service/{name}/start", StartServiceHandler).Methods("POST")
	r.HandleFunc("/api/service/{name}/stop", StopServiceHandler).Methods("POST")
	r.HandleFunc("/api/volume", GetVolumeHandler).Methods("GET")
	r.HandleFunc("/api/volume", SetVolumeHandler).Methods("POST")
	r.HandleFunc("/api/led", GetLEDHandler).Methods("GET")
	r.HandleFunc("/api/led", SetLEDHandler).Methods("POST")

	// 前端页面路由
	r.HandleFunc("/", IndexHandler)

	// 静态文件路由
	//r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("/Users/luopeng/GolandProjects/projectTest/aku/static"))))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// 启动服务器
	port := "8080"
	fmt.Printf("服务器正在运行，访问 http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	//tmpl := template.Must(template.ParseFiles("aku/index.html"))
	tmpl := template.Must(template.ParseFiles("index.html"))
	tmpl.Execute(w, services)
}

func GetServicesHandler(w http.ResponseWriter, r *http.Request) {
	muxLock.Lock()
	defer muxLock.Unlock()

	var result []ServiceStatus
	for _, service := range services {
		result = append(result, serviceStatusMap[service.Name])
	}
	json.NewEncoder(w).Encode(result)
}

func StartServiceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceName := vars["name"]

	// 执行启动命令
	cmd := exec.Command("systemctl", "start", serviceName)
	err := cmd.Run()
	if err != nil {
		http.Error(w, fmt.Sprintf("启动服务失败: %v", err), http.StatusInternalServerError)
		return
	}

	// 更新服务状态
	updateServiceStatus(serviceName)
	json.NewEncoder(w).Encode(map[string]string{"message": "服务已启动"})
}

func StopServiceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceName := vars["name"]

	// 执行停止命令
	cmd := exec.Command("systemctl", "stop", serviceName)
	err := cmd.Run()
	if err != nil {
		http.Error(w, fmt.Sprintf("停止服务失败: %v", err), http.StatusInternalServerError)
		return
	}

	// 更新服务状态
	updateServiceStatus(serviceName)
	json.NewEncoder(w).Encode(map[string]string{"message": "服务已停止"})
}

func monitorServices() {
	for {
		for _, service := range services {
			updateServiceStatus(service.Name)
		}
		time.Sleep(1 * time.Minute)
	}
}

func updateServiceStatus(serviceName string) {
	cmd := exec.Command("systemctl", "is-active", serviceName)
	output, err := cmd.CombinedOutput()
	status := strings.TrimSpace(string(output))
	if err != nil {
		status = "stop"
	}

	muxLock.Lock()
	tmp := serviceStatusMap[serviceName]
	tmp.Status = status
	serviceStatusMap[serviceName] = tmp
	muxLock.Unlock()
}

// 音量控制
func GetVolumeHandler(w http.ResponseWriter, r *http.Request) {
	// 执行获取音量的命令
	cmd := exec.Command("sh", "-c", "amixer get \"Power Amplifier\" | awk '/Mono:/ {print $2}'")
	output, err := cmd.CombinedOutput()
	if err != nil {
		http.Error(w, fmt.Sprintf("获取音量失败: %v", err), http.StatusInternalServerError)
		return
	}

	// 解析音量值
	volume := strings.TrimSpace(string(output))
	volume = strings.TrimSuffix(volume, "%")

	// 确保音量在1-63的范围内
	volumeInt, err := strconv.Atoi(volume)
	if err != nil {
		http.Error(w, fmt.Sprintf("解析音量失败: %v", err), http.StatusInternalServerError)
		return
	}
	volumeInt = int(math.Min(math.Max(float64(volumeInt), 1), 63))

	json.NewEncoder(w).Encode(volumeInt)
}

func SetVolumeHandler(w http.ResponseWriter, r *http.Request) {
	var volume struct {
		Volume int `json:"volume"`
	}
	if err := json.NewDecoder(r.Body).Decode(&volume); err != nil {
		http.Error(w, "请求解析失败", http.StatusBadRequest)
		return
	}

	// 确保音量在1-63的范围内
	volume.Volume = int(math.Min(math.Max(float64(volume.Volume), 1), 63))

	// 执行设置音量的命令
	cmd := exec.Command("sh", "-c", fmt.Sprintf("amixer set \"Power Amplifier\" %d", volume.Volume))
	err := cmd.Run()
	if err != nil {
		http.Error(w, fmt.Sprintf("设置音量失败: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "音量已更新"})
}

// LED灯控制
func GetLEDHandler(w http.ResponseWriter, r *http.Request) {
	// 读取LED状态文件
	content, err := os.ReadFile("/sys/class/leds/aku-logo/brightness")
	if err != nil {
		http.Error(w, fmt.Sprintf("读取LED状态失败: %v", err), http.StatusInternalServerError)
		return
	}

	// 解析LED状态
	state := string(content)
	state = strings.TrimSpace(state)

	// 返回LED状态
	json.NewEncoder(w).Encode(state == "1")
}

func SetLEDHandler(w http.ResponseWriter, r *http.Request) {
	var led struct {
		State bool `json:"state"`
	}
	if err := json.NewDecoder(r.Body).Decode(&led); err != nil {
		http.Error(w, "请求解析失败", http.StatusBadRequest)
		return
	}

	// 写入LED状态文件
	var stateStr string
	if led.State {
		stateStr = "1"
	} else {
		stateStr = "0"
	}

	err := os.WriteFile("/sys/class/leds/aku-logo/brightness", []byte(stateStr), 0666)
	if err != nil {
		http.Error(w, fmt.Sprintf("设置LED状态失败: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"state":   led.State,
	})
}
