import { toast as sonnerToast } from 'vue-sonner'
import { h } from 'vue'
import ErrorToast from '@/components/ui/sonner/ErrorToast.vue'

export const toast = {
    error(message: string) {
        sonnerToast.custom(h(ErrorToast, { message }), {
            duration: 8000,
        })
    },

    success(message: string) {
        sonnerToast.success(message)
    },

    info(message: string) {
        sonnerToast(message)
    },
}
