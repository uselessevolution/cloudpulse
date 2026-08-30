import { apiClient } from "./client";

import type { Service } from "../types/service";

export async function getServices(): Promise<Service[]> {
    const response =
        await apiClient.get<Service[]>(
            "/api/services/",
        );

    return response.data;
}