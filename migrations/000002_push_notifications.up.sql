CREATE TABLE IF NOT EXISTS push_notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid_v7(),
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    target_type VARCHAR(20) NOT NULL DEFAULT 'all',
    target_tenant_id UUID,
    target_user_ids UUID[],
    buttons JSONB,
    scheduled_at TIMESTAMP,
    sent_at TIMESTAMP,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    created_by UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_push_notifications_status ON push_notifications (status);

CREATE INDEX idx_push_notifications_scheduled_at ON push_notifications (scheduled_at)
WHERE
    status = 'scheduled';

CREATE INDEX idx_push_notifications_created_by ON push_notifications (created_by);

CREATE INDEX idx_push_notifications_target_tenant ON push_notifications (target_tenant_id)
WHERE
    target_tenant_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS push_notification_recipients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid_v7 (),
    push_notification_id UUID NOT NULL REFERENCES push_notifications (id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    delivered_at TIMESTAMP,
    read_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_push_recipients_notification ON push_notification_recipients (push_notification_id);

CREATE INDEX idx_push_recipients_user ON push_notification_recipients (user_id);

CREATE UNIQUE INDEX idx_push_recipients_unique ON push_notification_recipients (push_notification_id, user_id);