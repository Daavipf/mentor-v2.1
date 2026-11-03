-- 000003_add_time_metrics.sql
-- Colunas para acompanhamento e métricas de tempo

ALTER TABLE questions_on_exams
  ADD COLUMN IF NOT EXISTS started_at TIMESTAMP WITH TIME ZONE,
  ADD COLUMN IF NOT EXISTS finished_at TIMESTAMP WITH TIME ZONE
;