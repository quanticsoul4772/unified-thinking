# Monitoring and Operations Specification

## Overview

This document specifies monitoring, alerting, and operational excellence requirements for the unified-thinking MCP server in production environments.

## Monitoring Strategy

### Metrics Collection

The server exposes comprehensive metrics via the `get-metrics` tool:

```json
{
  "uptime_seconds": 3600,
  "total_thoughts": 1250,
  "total_branches": 45,
  "cache_hit_rate": 0.87,
  "storage_type": "sqlite",
  "memory_usage_mb": 128.5,
  "active_sessions": 3,
  "context_bridge": {
    "total_matches": 156,
    "cache_hits": 112,
    "cache_misses": 44,
    "error_count": 2,
    "timeout_count": 1,
    "latency_p50_ms": 45,
    "latency_p95_ms": 120,
    "latency_p99_ms": 250
  }
}
```

### Key Performance Indicators (KPIs)

#### 1. Availability Metrics

**Uptime**: Server availability percentage
- **Target**: 99.9% (8.76 hours downtime/year)
- **Measurement**: `uptime_seconds` / `total_time`
- **Alert Threshold**: < 99.5% over 24h period

**Response Time**: Time to process tool requests
- **Target**: P95 < 500ms, P99 < 1000ms
- **Measurement**: `context_bridge.latency_p95_ms`
- **Alert Threshold**: P95 > 1000ms or P99 > 2000ms

#### 2. Reliability Metrics

**Error Rate**: Failed requests per total requests
- **Target**: < 0.1% (1 in 1000 requests)
- **Measurement**: `error_count` / `total_requests`
- **Alert Threshold**: > 1% over 5min period

**Storage Persistence**: Data loss incidents
- **Target**: Zero data loss with SQLite backend
- **Measurement**: Failed writes to database
- **Alert Threshold**: Any write failure

**Context Bridge Timeouts**: Enrichment timeouts
- **Target**: < 5% timeout rate
- **Measurement**: `context_bridge.timeout_count` / `context_bridge.total_matches`
- **Alert Threshold**: > 10% over 10min period

#### 3. Performance Metrics

**Cache Hit Rate**: Percentage of cache hits
- **Target**: > 80% for frequently accessed data
- **Measurement**: `cache_hits` / (`cache_hits` + `cache_misses`)
- **Alert Threshold**: < 60% sustained over 30min

**Memory Usage**: Server memory consumption
- **Target**: < 512 MB under normal load
- **Measurement**: `memory_usage_mb`
- **Alert Threshold**: > 1024 MB (1 GB)

**Thought Processing Rate**: Thoughts processed per minute
- **Target**: > 100 thoughts/min under load
- **Measurement**: `delta(total_thoughts)` / `delta(time_seconds)` * 60
- **Alert Threshold**: < 50 thoughts/min during active usage

## Operational Monitoring

### Log Levels

#### DEBUG Level (Development Only)
```
[DEBUG] SQLite storage initialized successfully
[DEBUG] Context bridge matched 3 similar trajectories
[DEBUG] Embeddings enabled with provider: voyage
```

**When to Enable**: Development, troubleshooting, performance tuning
**Environment Variable**: `DEBUG=true`

#### INFO Level (Production Default)
```
[INFO] Server started on stdio transport
[INFO] Episodic memory handler initialized
[INFO] Warmed cache with 1000 thoughts
```

**When to Enable**: Always in production
**Use Case**: Normal operational visibility

#### WARNING Level
```
[WARN] Failed to warm cache: database locked
[WARN] Context bridge timeout exceeded (2s)
[WARN] Embedding generation failed, falling back to concept similarity
```

**Alert Criteria**: > 10 warnings/min
**Action**: Investigate underlying cause

#### ERROR Level
```
[ERROR] Failed to initialize embeddings: API key invalid
[ERROR] SQLite write failed: disk full
[ERROR] Probabilistic update failed: invalid likelihood values
```

**Alert Criteria**: Any ERROR logged
**Action**: Immediate investigation required

### Health Checks

#### Liveness Probe
**Endpoint**: Check server process is running
**Frequency**: Every 10 seconds
**Timeout**: 5 seconds
**Failure Action**: Restart server process

**Implementation**:
```bash
# Check if process is responding
ps aux | grep unified-thinking | grep -v grep
```

#### Readiness Probe
**Endpoint**: Check server can handle requests
**Frequency**: Every 30 seconds
**Timeout**: 10 seconds
**Failure Action**: Remove from load balancer

**Implementation**:
```json
// Call get-metrics tool and verify response
{
  "tool": "get-metrics",
  "expected_fields": ["uptime_seconds", "total_thoughts", "storage_type"]
}
```

### Storage Health

#### SQLite Database Checks

**Database Size Monitoring**:
```sql
SELECT page_count * page_size as size_bytes FROM pragma_page_count(), pragma_page_size();
```

**Alert Thresholds**:
- WARNING: Database > 1 GB
- CRITICAL: Database > 10 GB or > 80% disk space

**Integrity Check** (Daily):
```sql
PRAGMA integrity_check;
```

**Alert**: Any result other than "ok"

**WAL File Growth**:
```bash
ls -lh $SQLITE_PATH-wal
```

**Alert**: WAL file > 100 MB (indicates checkpointing issues)

#### Cache Performance

**Warm Cache Success Rate**:
```
warmed_thoughts / WARM_CACHE_LIMIT
```

**Target**: > 95% of configured limit
**Alert**: < 80% (indicates database issues or memory pressure)

**Cache Eviction Rate**:
```
cache_evictions / cache_total_operations
```

**Target**: < 10% eviction rate
**Alert**: > 25% (indicates insufficient cache size)

## Alerting Rules

### Critical Alerts (Page Immediately)

1. **Server Crash**
   - Condition: Process terminated unexpectedly
   - Impact: Complete service unavailability
   - Response: Immediate restart and investigation

2. **Database Corruption**
   - Condition: `PRAGMA integrity_check` fails
   - Impact: Potential data loss
   - Response: Restore from backup, investigate cause

3. **Out of Memory**
   - Condition: `memory_usage_mb` > 2048 MB
   - Impact: Server crash imminent
   - Response: Restart server, investigate memory leak

4. **Disk Full**
   - Condition: Available disk space < 5%
   - Impact: SQLite writes failing
   - Response: Free disk space, rotate logs

### High Priority Alerts (Resolve Within 1 Hour)

1. **High Error Rate**
   - Condition: Error rate > 1% for 5 minutes
   - Impact: User experience degradation
   - Response: Check logs, identify root cause

2. **Slow Response Time**
   - Condition: P95 latency > 1000ms for 10 minutes
   - Impact: Poor user experience
   - Response: Check database performance, CPU usage

3. **Context Bridge Failures**
   - Condition: Timeout rate > 10% for 10 minutes
   - Impact: Degraded context enrichment
   - Response: Check embedding API, network connectivity

4. **Cache Performance Degradation**
   - Condition: Cache hit rate < 60% for 30 minutes
   - Impact: Increased database load, slower responses
   - Response: Check cache configuration, database performance

### Medium Priority Alerts (Resolve Within 4 Hours)

1. **Warning Log Spike**
   - Condition: > 50 warnings in 10 minutes
   - Impact: Potential issues developing
   - Response: Review warning patterns, address causes

2. **Database Size Growth**
   - Condition: Database > 1 GB
   - Impact: Performance degradation potential
   - Response: Archive old data, optimize queries

3. **Low Cache Warm Success**
   - Condition: Warm cache < 80% for 1 hour
   - Impact: Suboptimal performance
   - Response: Increase `WARM_CACHE_LIMIT` or investigate database issues

## Operational Runbooks

### Runbook: Server Crash Recovery

**Symptoms**: Process terminated, no response to health checks

**Investigation**:
1. Check system logs: `journalctl -u unified-thinking -n 100`
2. Check available memory: `free -h`
3. Check disk space: `df -h`
4. Review recent error logs in working directory

**Resolution**:
1. If out of memory: Increase server memory allocation
2. If disk full: Free disk space, rotate logs
3. If crash in SQLite: Restore from backup
4. Restart server process

**Prevention**:
- Monitor memory usage trends
- Implement log rotation
- Regular database backups

### Runbook: Database Corruption

**Symptoms**: `PRAGMA integrity_check` fails, SQLite errors in logs

**Investigation**:
1. Stop server to prevent further corruption
2. Create backup of corrupted database: `cp thoughts.db thoughts.db.corrupt`
3. Run integrity check manually:
   ```sql
   sqlite3 thoughts.db "PRAGMA integrity_check"
   ```

**Resolution**:
1. Attempt recovery:
   ```sql
   sqlite3 thoughts.db ".recover" > recovered.sql
   sqlite3 new_thoughts.db < recovered.sql
   ```
2. If recovery fails: Restore from latest backup
3. Verify restored database: `PRAGMA integrity_check`
4. Restart server with restored database

**Prevention**:
- Enable daily backups
- Monitor WAL file size
- Use `PRAGMA wal_checkpoint(TRUNCATE)` regularly

### Runbook: High Memory Usage

**Symptoms**: `memory_usage_mb` increasing over time, approaching limits

**Investigation**:
1. Check current memory usage: `get-metrics` tool
2. Review recent thought/branch creation patterns
3. Check cache size and eviction rate
4. Look for memory leaks in error logs

**Resolution**:
1. Restart server to clear memory (short-term)
2. Reduce `WARM_CACHE_LIMIT` if cache is too large
3. Archive old thoughts/branches to reduce in-memory data
4. If memory leak suspected: Collect profiling data, report issue

**Prevention**:
- Set appropriate `WARM_CACHE_LIMIT` based on available memory
- Implement periodic cache cleanup
- Monitor memory usage trends

### Runbook: Slow Response Times

**Symptoms**: P95/P99 latency increasing, user complaints

**Investigation**:
1. Check context bridge metrics: `context_bridge.latency_p95_ms`
2. Review database query performance
3. Check system CPU and I/O utilization
4. Verify embedding API response times (if enabled)

**Resolution**:
1. If database slow: Run `PRAGMA optimize` and `VACUUM`
2. If cache cold: Increase `WARM_CACHE_LIMIT`
3. If CPU bound: Reduce concurrent requests or scale vertically
4. If embedding API slow: Implement longer timeouts or disable embeddings temporarily

**Prevention**:
- Regular database maintenance (VACUUM, ANALYZE)
- Monitor database query performance
- Implement query result caching
- Set appropriate timeouts for external APIs

## Capacity Planning

### Resource Requirements

#### Minimum Configuration (Development)
- **CPU**: 1 core
- **Memory**: 256 MB
- **Disk**: 1 GB
- **Expected Load**: < 10 requests/min

#### Production Configuration (Standard)
- **CPU**: 2-4 cores
- **Memory**: 512 MB - 1 GB
- **Disk**: 10 GB (with log rotation)
- **Expected Load**: 100-500 requests/min

#### Production Configuration (High Load)
- **CPU**: 4-8 cores
- **Memory**: 2 GB - 4 GB
- **Disk**: 50 GB (with archival)
- **Expected Load**: 500-2000 requests/min

### Scaling Considerations

**Vertical Scaling** (Recommended for single-user):
- Increase memory for larger cache
- Increase CPU for faster processing
- Use SSD for SQLite database

**Horizontal Scaling** (Multi-user deployment):
- Run multiple server instances
- Shared SQLite database (read replicas)
- Load balancer with health checks
- Consider PostgreSQL for true multi-writer

### Data Retention

**Thought/Branch Data**:
- Retention: 90 days for active reasoning
- Archival: Export to cold storage after 90 days
- Deletion: User-initiated or after 1 year

**Episodic Memory**:
- Retention: Indefinite (learning data)
- Archival: Compress old trajectories (> 6 months)
- Deletion: Never (unless privacy required)

**Logs**:
- Retention: 30 days in hot storage
- Archival: Compress and archive for 1 year
- Deletion: After 1 year

## Security Monitoring

### Audit Events

**Track the following events**:
1. Server start/stop
2. Configuration changes
3. Database schema migrations
4. Failed authentication attempts (if auth added)
5. Suspicious request patterns (rate limiting triggers)

**Log Format**:
```json
{
  "timestamp": "2025-01-20T10:30:00Z",
  "event": "server_start",
  "user": "system",
  "details": {"storage_type": "sqlite", "embeddings_enabled": true}
}
```

### Security Alerts

1. **Unusual Request Patterns**
   - Condition: > 1000 requests/min from single source
   - Action: Investigate potential DoS attack

2. **SQL Injection Attempts**
   - Condition: Suspicious SQL patterns in request params
   - Action: Block source, review injection prevention

3. **Unauthorized File Access**
   - Condition: Attempts to read files outside data directory
   - Action: Log and block, review access controls

## Observability Dashboard

### Recommended Metrics Dashboard

**Section 1: Health Overview**
- Server uptime (gauge)
- Request rate (line chart, requests/min)
- Error rate (line chart, errors/min)
- Current memory usage (gauge, MB)

**Section 2: Performance**
- Response time P50/P95/P99 (line chart, ms)
- Cache hit rate (gauge, %)
- Database query time (histogram, ms)
- Context bridge latency (line chart, ms)

**Section 3: Resource Utilization**
- CPU usage (line chart, %)
- Memory usage (line chart, MB)
- Disk usage (gauge, GB)
- Database size (gauge, MB)

**Section 4: Business Metrics**
- Thoughts created (counter)
- Branches created (counter)
- Active reasoning sessions (gauge)
- Episodic memory trajectory count (counter)

## Conclusion

This monitoring and operations specification provides a comprehensive framework for production deployment of the unified-thinking MCP server. Key principles:

1. **Proactive Monitoring**: Detect issues before they impact users
2. **Clear Alerting**: Prioritize alerts by severity and required response time
3. **Actionable Runbooks**: Provide step-by-step recovery procedures
4. **Capacity Planning**: Right-size resources for expected load
5. **Security Awareness**: Monitor for suspicious patterns

Regular review and updates of this specification ensure the server maintains high availability, performance, and reliability in production environments.
