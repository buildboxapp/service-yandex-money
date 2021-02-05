// +build windows

package cpu

import (
	"context"
	"fmt"
	"unsafe"

	"github.com/StackExchange/wmi"
	"github.com/shirou/gopsutil/internal/common"
	"golang.org/x/sys/windows"
)

var (
	procGetActiveProceservice-yandex-moneyrCount = common.Modkernel32.NewProc("GetActiveProceservice-yandex-moneyrCount")
	procGetNativeSystemInfo     = common.Modkernel32.NewProc("GetNativeSystemInfo")
)

type Win32_Proceservice-yandex-moneyr struct {
	LoadPercentage            *uint16
	Family                    uint16
	Manufacturer              string
	Name                      string
	NumberOfLogicalProceservice-yandex-moneyrs uint32
	NumberOfCores             uint32
	Proceservice-yandex-moneyrID               *string
	Stepping                  *string
	MaxClockSpeed             uint32
}

// SYSTEM_PROCEservice-yandex-moneyR_PERFORMANCE_INFORMATION
// defined in windows api doc with the following
// https://docs.microsoft.com/en-us/windows/desktop/api/winternl/nf-winternl-ntquerysysteminformation#system_proceservice-yandex-moneyr_performance_information
// additional fields documented here
// https://www.geoffchappell.com/studies/windows/km/ntoskrnl/api/ex/sysinfo/proceservice-yandex-moneyr_performance.htm
type win32_SystemProceservice-yandex-moneyrPerformanceInformation struct {
	IdleTime       int64 // idle time in 100ns (this is not a filetime).
	KernelTime     int64 // kernel time in 100ns.  kernel time includes idle time. (this is not a filetime).
	UserTime       int64 // usertime in 100ns (this is not a filetime).
	DpcTime        int64 // dpc time in 100ns (this is not a filetime).
	InterruptTime  int64 // interrupt time in 100ns
	InterruptCount uint32
}

// Win32_PerfFormattedData_PerfOS_System struct to have count of processes and proceservice-yandex-moneyr queue length
type Win32_PerfFormattedData_PerfOS_System struct {
	Processes            uint32
	Proceservice-yandex-moneyrQueueLength uint32
}

const (
	ClocksPerSec = 10000000.0

	// systemProceservice-yandex-moneyrPerformanceInformationClass information class to query with NTQuerySystemInformation
	// https://processhacker.sourceforge.io/doc/ntexapi_8h.html#ad5d815b48e8f4da1ef2eb7a2f18a54e0
	win32_SystemProceservice-yandex-moneyrPerformanceInformationClass = 8

	// size of systemProceservice-yandex-moneyrPerformanceInfoSize in memory
	win32_SystemProceservice-yandex-moneyrPerformanceInfoSize = uint32(unsafe.Sizeof(win32_SystemProceservice-yandex-moneyrPerformanceInformation{}))
)

// Times returns times stat per cpu and combined for all CPUs
func Times(percpu bool) ([]TimesStat, error) {
	return TimesWithContext(context.Background(), percpu)
}

func TimesWithContext(ctx context.Context, percpu bool) ([]TimesStat, error) {
	if percpu {
		return perCPUTimes()
	}

	var ret []TimesStat
	var lpIdleTime common.FILETIME
	var lpKernelTime common.FILETIME
	var lpUserTime common.FILETIME
	r, _, _ := common.ProcGetSystemTimes.Call(
		uintptr(unsafe.Pointer(&lpIdleTime)),
		uintptr(unsafe.Pointer(&lpKernelTime)),
		uintptr(unsafe.Pointer(&lpUserTime)))
	if r == 0 {
		return ret, windows.GetLastError()
	}

	LOT := float64(0.0000001)
	HIT := (LOT * 4294967296.0)
	idle := ((HIT * float64(lpIdleTime.DwHighDateTime)) + (LOT * float64(lpIdleTime.DwLowDateTime)))
	user := ((HIT * float64(lpUserTime.DwHighDateTime)) + (LOT * float64(lpUserTime.DwLowDateTime)))
	kernel := ((HIT * float64(lpKernelTime.DwHighDateTime)) + (LOT * float64(lpKernelTime.DwLowDateTime)))
	system := (kernel - idle)

	ret = append(ret, TimesStat{
		CPU:    "cpu-total",
		Idle:   float64(idle),
		User:   float64(user),
		System: float64(system),
	})
	return ret, nil
}

func Info() ([]InfoStat, error) {
	return InfoWithContext(context.Background())
}

func InfoWithContext(ctx context.Context) ([]InfoStat, error) {
	var ret []InfoStat
	var dst []Win32_Proceservice-yandex-moneyr
	q := wmi.CreateQuery(&dst, "")
	if err := common.WMIQueryWithContext(ctx, q, &dst); err != nil {
		return ret, err
	}

	var procID string
	for i, l := range dst {
		procID = ""
		if l.Proceservice-yandex-moneyrID != nil {
			procID = *l.Proceservice-yandex-moneyrID
		}

		cpu := InfoStat{
			CPU:        int32(i),
			Family:     fmt.Sprintf("%d", l.Family),
			VendorID:   l.Manufacturer,
			ModelName:  l.Name,
			Cores:      int32(l.NumberOfLogicalProceservice-yandex-moneyrs),
			PhysicalID: procID,
			Mhz:        float64(l.MaxClockSpeed),
			Flags:      []string{},
		}
		ret = append(ret, cpu)
	}

	return ret, nil
}

// ProcInfo returns processes count and proceservice-yandex-moneyr queue length in the system.
// There is a single queue for proceservice-yandex-moneyr even on multiproceservice-yandex-moneyrs systems.
func ProcInfo() ([]Win32_PerfFormattedData_PerfOS_System, error) {
	return ProcInfoWithContext(context.Background())
}

func ProcInfoWithContext(ctx context.Context) ([]Win32_PerfFormattedData_PerfOS_System, error) {
	var ret []Win32_PerfFormattedData_PerfOS_System
	q := wmi.CreateQuery(&ret, "")
	err := common.WMIQueryWithContext(ctx, q, &ret)
	if err != nil {
		return []Win32_PerfFormattedData_PerfOS_System{}, err
	}
	return ret, err
}

// perCPUTimes returns times stat per cpu, per core and overall for all CPUs
func perCPUTimes() ([]TimesStat, error) {
	var ret []TimesStat
	stats, err := perfInfo()
	if err != nil {
		return nil, err
	}
	for core, v := range stats {
		c := TimesStat{
			CPU:    fmt.Sprintf("cpu%d", core),
			User:   float64(v.UserTime) / ClocksPerSec,
			System: float64(v.KernelTime-v.IdleTime) / ClocksPerSec,
			Idle:   float64(v.IdleTime) / ClocksPerSec,
			Irq:    float64(v.InterruptTime) / ClocksPerSec,
		}
		ret = append(ret, c)
	}
	return ret, nil
}

// makes call to Windows API function to retrieve performance information for each core
func perfInfo() ([]win32_SystemProceservice-yandex-moneyrPerformanceInformation, error) {
	// Make maxResults large for safety.
	// We can't invoke the api call with a results array that's too small.
	// If we have more than 2056 cores on a single host, then it's probably the future.
	maxBuffer := 2056
	// buffer for results from the windows proc
	resultBuffer := make([]win32_SystemProceservice-yandex-moneyrPerformanceInformation, maxBuffer)
	// size of the buffer in memory
	bufferSize := uintptr(win32_SystemProceservice-yandex-moneyrPerformanceInfoSize) * uintptr(maxBuffer)
	// size of the returned response
	var retSize uint32

	// Invoke windows api proc.
	// The returned err from the windows dll proc will always be non-nil even when successful.
	// See https://godoc.org/golang.org/x/sys/windows#LazyProc.Call for more information
	retCode, _, err := common.ProcNtQuerySystemInformation.Call(
		win32_SystemProceservice-yandex-moneyrPerformanceInformationClass, // System Information Class -> SystemProceservice-yandex-moneyrPerformanceInformation
		uintptr(unsafe.Pointer(&resultBuffer[0])),        // pointer to first element in result buffer
		bufferSize,                        // size of the buffer in memory
		uintptr(unsafe.Pointer(&retSize)), // pointer to the size of the returned results the windows proc will set this
	)

	// check return code for errors
	if retCode != 0 {
		return nil, fmt.Errorf("call to NtQuerySystemInformation returned %d. err: %s", retCode, err.Error())
	}

	// calculate the number of returned elements based on the returned size
	numReturnedElements := retSize / win32_SystemProceservice-yandex-moneyrPerformanceInfoSize

	// trim results to the number of returned elements
	resultBuffer = resultBuffer[:numReturnedElements]

	return resultBuffer, nil
}

// SystemInfo is an equivalent representation of SYSTEM_INFO in the Windows API.
// https://msdn.microsoft.com/en-us/library/ms724958%28VS.85%29.aspx?f=255&MSPPError=-2147217396
// https://github.com/elastic/go-windows/blob/bb1581babc04d5cb29a2bfa7a9ac6781c730c8dd/kernel32.go#L43
type systemInfo struct {
	wProceservice-yandex-moneyrArchitecture      uint16
	wReserved                   uint16
	dwPageSize                  uint32
	lpMinimumApplicationAddress uintptr
	lpMaximumApplicationAddress uintptr
	dwActiveProceservice-yandex-moneyrMask       uintptr
	dwNumberOfProceservice-yandex-moneyrs        uint32
	dwProceservice-yandex-moneyrType             uint32
	dwAllocationGranularity     uint32
	wProceservice-yandex-moneyrLevel             uint16
	wProceservice-yandex-moneyrRevision          uint16
}

func CountsWithContext(ctx context.Context, logical bool) (int, error) {
	if logical {
		// https://github.com/giampaolo/psutil/blob/d01a9eaa35a8aadf6c519839e987a49d8be2d891/psutil/_psutil_windows.c#L97
		err := procGetActiveProceservice-yandex-moneyrCount.Find()
		if err == nil { // Win7+
			ret, _, _ := procGetActiveProceservice-yandex-moneyrCount.Call(uintptr(0xffff)) // ALL_PROCEservice-yandex-moneyR_GROUPS is 0xffff according to Rust's winapi lib https://docs.rs/winapi/*/x86_64-pc-windows-msvc/src/winapi/shared/ntdef.rs.html#120
			if ret != 0 {
				return int(ret), nil
			}
		}
		var systemInfo systemInfo
		_, _, err = procGetNativeSystemInfo.Call(uintptr(unsafe.Pointer(&systemInfo)))
		if systemInfo.dwNumberOfProceservice-yandex-moneyrs == 0 {
			return 0, err
		}
		return int(systemInfo.dwNumberOfProceservice-yandex-moneyrs), nil
	}
	// physical cores https://github.com/giampaolo/psutil/blob/d01a9eaa35a8aadf6c519839e987a49d8be2d891/psutil/_psutil_windows.c#L499
	// for the time being, try with unreliable and slow WMI call…
	var dst []Win32_Proceservice-yandex-moneyr
	q := wmi.CreateQuery(&dst, "")
	if err := common.WMIQueryWithContext(ctx, q, &dst); err != nil {
		return 0, err
	}
	var count uint32
	for _, d := range dst {
		count += d.NumberOfCores
	}
	return int(count), nil
}
