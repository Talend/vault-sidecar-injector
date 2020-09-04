# Metrics

Vault Sidecar Injector exposes a Prometheus endpoint at `/metrics` on port `metricsPort` (default: 9000).

Following collectors are available:

<details>
<summary>
Process Collector
</summary>

- process_cpu_seconds_total
- process_virtual_memory_bytes
- process_start_time_seconds
- process_open_fds
- process_max_fds
</details>

<details>
<summary>
Go Collector
</summary>

- go_goroutines
- go_threads
- go_gc_duration_seconds
- go_info
- go_memstats_alloc_bytes
- go_memstats_heap_alloc_bytes
- go_memstats_alloc_bytes_total
- go_memstats_sys_bytes
- go_memstats_lookups_total
- ... [full list here](https://github.com/prometheus/client_golang/blob/master/prometheus/go_collector.go)
</details>

![Grafana dashboard](grafana-vault-sidecar-injector.png)
