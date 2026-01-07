/**
 * StatusBar Component
 * Displays context size and model information
 */

const StatusBar = {
    name: 'StatusBar',

    props: {
        contextSize: {
            type: Number,
            default: 0
        },
        models: {
            type: Object,
            default: () => ({})
        },
        selectedAgent: {
            type: String,
            default: 'unknown'
        }
    },

    setup(props) {
        const formatContextSize = Vue.computed(() => {
            const size = props.contextSize;
            if (size >= 1000000) {
                return `${(size / 1000000).toFixed(2)}M`;
            } else if (size >= 1000) {
                return `${(size / 1000).toFixed(2)}K`;
            }
            return size.toString();
        });

        const chatModel = Vue.computed(() => {
            return props.models.chat_model || 'N/A';
        });

        const toolsModel = Vue.computed(() => {
            return props.models.tools_model || 'N/A';
        });

        const ragModel = Vue.computed(() => {
            return props.models.rag_model || 'N/A';
        });

        return {
            formatContextSize,
            chatModel,
            toolsModel,
            ragModel
        };
    },

    template: `
        <div class="status-bar">
            <div class="status-item">
                <span class="status-label">Agent:</span>
                <span class="status-value">{{ selectedAgent }}</span>
            </div>
            <div class="status-item">
                <span class="status-label">Context Size:</span>
                <span class="status-value">{{ formatContextSize }}</span>
            </div>
            <div class="status-item" v-if="chatModel !== 'N/A'">
                <span class="status-label">Chat Model:</span>
                <span class="status-value">{{ chatModel }}</span>
            </div>
            <div class="status-item" v-if="toolsModel !== 'N/A'">
                <span class="status-label">Tools:</span>
                <span class="status-value">{{ toolsModel }}</span>
            </div>
            <div class="status-item" v-if="ragModel !== 'N/A'">
                <span class="status-label">RAG:</span>
                <span class="status-value">{{ ragModel }}</span>
            </div>
        </div>
    `
};

// Register component globally
window.StatusBar = StatusBar;
