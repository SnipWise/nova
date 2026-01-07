/**
 * Modal Component
 * Reusable modal dialog for displaying information and confirmations
 */

const Modal = {
    name: 'Modal',

    props: {
        show: {
            type: Boolean,
            required: true
        },
        title: {
            type: String,
            default: 'Dialog'
        },
        type: {
            type: String,
            default: 'info', // 'info', 'confirm', 'list'
            validator: (value) => ['info', 'confirm', 'list'].includes(value)
        },
        width: {
            type: String,
            default: '600px'
        }
    },

    emits: ['close', 'confirm', 'cancel'],

    setup(props, { emit }) {
        const handleClose = () => {
            emit('close');
        };

        const handleConfirm = () => {
            emit('confirm');
        };

        const handleCancel = () => {
            emit('cancel');
            emit('close');
        };

        const handleBackdropClick = (event) => {
            if (event.target === event.currentTarget) {
                handleClose();
            }
        };

        return {
            handleClose,
            handleConfirm,
            handleCancel,
            handleBackdropClick
        };
    },

    template: `
        <transition name="modal">
            <div v-if="show" class="modal-backdrop" @click="handleBackdropClick">
                <div class="modal-container" :style="{ maxWidth: width }">
                    <div class="modal-header">
                        <h3>{{ title }}</h3>
                        <button class="modal-close" @click="handleClose">×</button>
                    </div>

                    <div class="modal-body">
                        <slot></slot>
                    </div>

                    <div class="modal-footer" v-if="type === 'confirm'">
                        <button class="modal-btn" @click="handleCancel">Cancel</button>
                        <button class="modal-btn modal-btn-danger" @click="handleConfirm">Confirm</button>
                    </div>

                    <div class="modal-footer" v-else>
                        <button class="modal-btn modal-btn-primary" @click="handleClose">Close</button>
                    </div>
                </div>
            </div>
        </transition>
    `
};

// Register component globally
window.Modal = Modal;
