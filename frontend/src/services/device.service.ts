import { api } from './api';

export interface Device {
    id: string;
    name: string;
    publicKey: string;
    encryptedMasterKey?: string;
    lastActive: string;
    userId: string;
}

export interface RegisterDeviceRequest {
    name: string;
    publicKey: string;
    encryptedMasterKey?: string;
}

export interface UpdateDeviceRequest {
    encryptedMasterKey?: string;
}

export interface RegisterWaitingDeviceRequest {
    name: string;
    publicKey: string;
}

export class DeviceService {
    static async getDevices(): Promise<Device[]> {
        return api.get<Device[]>('/devices');
    }

    static async registerDevice(request: RegisterDeviceRequest): Promise<Device> {
        return api.post<Device>('/devices', request);
    }

    static async updateDevice(deviceId: string, request: UpdateDeviceRequest): Promise<Device> {
        return api.put<Device>(`/devices/${deviceId}`, request);
    }

    static async deleteDevice(deviceId: string): Promise<void> {
        await api.delete(`/devices/${deviceId}`);
    }

    static async deleteAllDevices(): Promise<void> {
        await api.delete('/devices');
    }
}
