import { useQuery } from '@tanstack/vue-query';
import { DeviceService } from '@/services/device.service';
import { queryKeys } from './keys';

export function useDevices() {
    return useQuery({
        queryKey: queryKeys.devices,
        queryFn: () => DeviceService.getDevices(),
    });
}
