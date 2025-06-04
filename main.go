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
	"gopkg.in/yaml.v3"
)

type ServiceStatus struct {
	Name        string `json:"name"`
	Status      string `json:"status"`
	Description string `json:"description"`
}

// 城市结构体
type City struct {
	City     string `yaml:"city"`
	Code     string `yaml:"code"`
	Province string `yaml:"province"`
	URL      string `yaml:"url"`
}

// 省份结构体
type Province struct {
	Code   string `yaml:"code"`
	Name   string `yaml:"name"`
	URL    string `yaml:"url"`
	Cities []City `yaml:"-"` // 用于存储解析出的城市信息
}

// tqstation1.yaml 数据结构
var tqstation1 map[string]Province
var (
	services = []ServiceStatus{
		{Name: "xiaozhi", Description: "小智语音服务"},
		{Name: "sysboot", Description: "系统启动服务"},
		{Name: "showtime", Description: "时间显示服务"},
		{Name: "fm", Description: "网络收音机服务"},
		{Name: "ttyd", Description: "网页终端服务"},
		{Name: "shairport-sync", Description: "AirPlay服务"},
		{Name: "bluealsad-aplay", Description: "蓝牙服务"},
		{Name: "dlna", Description: "DLNA服务"},
		{Name: "alist", Description: "文件服务"},
		{Name: "akuweb", Description: "网页控制台服务"},
		{Name: "akuapp", Description: "APP后台服务"},
	}
	serviceStatusMap = make(map[string]ServiceStatus)
	muxLock          sync.Mutex
)

func main() {
	// 初始化服务状态
	for _, service := range services {
		serviceStatusMap[service.Name] = service
	}
	// 读取 tqstation1.yaml 文件
	if err := loadTQStationYAML(); err != nil {
		log.Fatalf("加载 tqstation1.yaml 失败: %v", err)
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
	// 添加城市数据相关的路由
	r.HandleFunc("/api/cities", GetCitiesHandler).Methods("GET")
	r.HandleFunc("/api/currentcity", GetCurrentCityHandler).Methods("GET")
	r.HandleFunc("/api/city", UpdateCityCodeHandler).Methods("POST")
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

// 加载 tqstation1.yaml 文件
func loadTQStationYAML() error {
	yamlFile, err := os.ReadFile("/etc/tqstation1.yaml")
	if err != nil {
		return err
	}

	// 定义一个临时结构来解析省份的基本信息和城市信息
	type RawProvince map[string]interface{}

	// 临时存储解析后的数据
	rawTqstation1 := make(map[string]RawProvince)
	err = yaml.Unmarshal(yamlFile, &rawTqstation1)
	if err != nil {
		return err
	}

	// 初始化 tqstation1
	tqstation1 = make(map[string]Province)

	// 遍历每个省份
	for provinceName, rawProvince := range rawTqstation1 {
		var province Province

		// 解析省份的基本信息
		if code, ok := rawProvince["code"].(string); ok {
			province.Code = code
		}
		if name, ok := rawProvince["name"].(string); ok {
			province.Name = name
		}
		if url, ok := rawProvince["url"].(string); ok {
			province.URL = url
		}
		// 解析城市信息
		for key, value := range rawProvince {
			// 跳过省份的基本信息字段
			if key == "code" || key == "name" || key == "url" {
				continue
			}

			// 检查是否是城市信息
			jsonData, _ := json.Marshal(value)
			var city City
			json.Unmarshal(jsonData, &city)
			province.Cities = append(province.Cities, city)
		}

		tqstation1[provinceName] = province
	}

	// 打印解析后的数据以检查
	//fmt.Println(tqstation1)
	return nil
}

// 获取所有城市
func GetCitiesHandler(w http.ResponseWriter, r *http.Request) {
	var cities []City
	for _, province := range tqstation1 {
		for _, city := range province.Cities {
			cities = append(cities, city)
		}
	}
	json.NewEncoder(w).Encode(cities)
}

// 获取当前城市配置
func GetCurrentCityHandler(w http.ResponseWriter, r *http.Request) {
	// 直接读取文件内容
	configData, err := os.ReadFile("/etc/akutq_city.conf")
	if err != nil {
		http.Error(w, fmt.Sprintf("获取城市配置失败: %v", err), http.StatusInternalServerError)
		return
	}

	configStr := string(configData)
	configStr = strings.TrimSpace(configStr)
	parts := strings.Split(configStr, "=")
	if len(parts) != 2 {
		http.Error(w, "城市配置格式错误", http.StatusInternalServerError)
		return
	}
	cityCode := parts[1]

	// 查找城市信息
	var cityInfo City
	for _, province := range tqstation1 {
		for _, city := range province.Cities {
			if city.Code == cityCode {
				cityInfo = city
				break
			}
		}
		if cityInfo.Code != "" {
			break
		}
	}

	if cityInfo.Code == "" {
		http.Error(w, "未找到城市信息", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"city":     cityInfo.City,
		"code":     cityInfo.Code,
		"province": cityInfo.Province,
	})
}

// 更新城市代码
func UpdateCityCodeHandler(w http.ResponseWriter, r *http.Request) {
	var city struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&city); err != nil {
		http.Error(w, "请求解析失败", http.StatusBadRequest)
		return
	}

	// 写入城市配置文件
	content := fmt.Sprintf("city_code=%s", city.Code)
	err := os.WriteFile("/etc/akutq_city.conf", []byte(content), 0644)
	if err != nil {
		http.Error(w, fmt.Sprintf("更新城市配置失败: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "城市已更新"})
}
