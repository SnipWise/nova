/**
 * OperationControls Component
 * Displays controls for validating/canceling pending operations (human-in-the-loop)
 */

const OperationControls = {
    name: 'OperationControls',

    props: {
        operation: {
            type: Object,
            required: true
        }
    },

    emits: ['validate', 'cancel'],

    setup(props, { emit }) {
        const handleValidate = () => {
            emit('validate', props.operation.operation_id);
        };

        const handleCancel = () => {
            emit('cancel', props.operation.operation_id);
        };

        const statusLabel = Vue.computed(() => {
            const status = props.operation.status || 'unknown';
            return status.charAt(0).toUpperCase() + status.slice(1);
        });

        const operationMessage = Vue.computed(() => {
            return props.operation.message || 'Operation pending...';
        });

        const operationId = Vue.computed(() => {
            return props.operation.operation_id || 'N/A';
        });

        const showControls = Vue.computed(() => {
            return props.operation.status === 'pending';
        });

        const statusIcon = Vue.computed(() => {
            switch (props.operation.status) {
                case 'pending': return '⏳';
                case 'completed': return '✅';
                case 'cancelled': return '❌';
                default: return '🔔';
            }
        });

        const statusClass = Vue.computed(() => {
            return `operation-controls operation-${props.operation.status || 'pending'}`;
        });

        return {
            handleValidate,
            handleCancel,
            statusLabel,
            operationMessage,
            operationId,
            showControls,
            statusIcon,
            statusClass
        };
    },

    template: `
        <div :class="statusClass">
            <h4>{{ statusIcon }} Tool Call Notification</h4>
            <p>{{ operationMessage }}</p>
            <div class="status-item">
                <span class="status-label">Status:</span>
                <span class="status-value">{{ statusLabel }}</span>
            </div>
            <div class="status-item">
                <span class="status-label">Operation ID:</span>
                <span class="status-value">{{ operationId }}</span>
            </div>
            <div class="operation-buttons" v-if="showControls">
                <button
                    class="success"
                    @click="handleValidate"
                >
                    ✓ Validate
                </button>
                <button
                    class="danger"
                    @click="handleCancel"
                >
                    ✗ Cancel
                </button>
            </div>
            <div v-else-if="operation.status === 'completed'" class="operation-result">
                ✅ Operation validated successfully
            </div>
            <div v-else-if="operation.status === 'cancelled'" class="operation-result">
                ❌ Operation cancelled
            </div>
        </div>
    `
};

// Register component globally
window.OperationControls = OperationControls;
