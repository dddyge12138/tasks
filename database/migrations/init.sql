-- 创建tasks表
drop table if exists tasks;
CREATE TABLE tasks (
    id BIGSERIAL PRIMARY KEY,
    task_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    status SMALLINT NOT NULL DEFAULT 1, -- 任务状态：1-待执行，2-执行中，3-已完成，4-已失败
    cron VARCHAR(100) DEFAULT NULL, -- cron表达式，为空表示一次性任务
    next_pending_time BIGINT default 0,
    params JSONB NOT NULL, -- 任务参数JSON数组
    cron_task_ids BIGINT[] DEFAULT NULL, -- 存储拆分后的子任务ID数组
    is_deleted SMALLINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_tasks_next_pending_time ON tasks(next_pending_time);
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_cron_task_ids ON tasks USING GIN(cron_task_ids);
create unique index idx_tasks_task_id on tasks(task_id);

-- 创建task_executions表
drop table if exists task_executions;
CREATE TABLE IF NOT EXISTS task_executions (
    id BIGSERIAL PRIMARY KEY,
    task_id BIGINT NOT NULL,
    execution_time TIMESTAMP WITH TIME ZONE NOT NULL,
    status SMALLINT NOT NULL, -- 执行状态：1-执行中，2-成功，3-失败
    error_message TEXT DEFAULT NULL,
    retry_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_task_executions_task_id ON task_executions(task_id);

-- 创建task_results表
drop table if exists task_results;
CREATE TABLE IF NOT EXISTS task_results (
    id BIGSERIAL PRIMARY KEY,
    task_id BIGINT NOT NULL,
    cron_task_id BIGINT NOT NULL,
    status SMALLINT NOT NULL, -- 状态：1-成功，2-失败
    result JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_task_results_task_id ON task_results(task_id);
CREATE INDEX idx_task_results_cron_task_id ON task_results(cron_task_id);

