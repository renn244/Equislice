# Adopt the existing deployment. Review the plan before applying.
import {
  to = azurerm_storage_account.equislice-sa
  id = local.storage_id
}
import {
  for_each = toset(["equirectangular", "equirectangular-slice"])
  to       = azurerm_storage_container.equislice-sa-containers[each.value]
  id       = "${local.storage_id}/blobServices/default/containers/${each.value}"
}
import {
  to = azurerm_storage_container.deployment
  id = "${local.storage_id}/blobServices/default/containers/${local.deployment_container_name}"
}
import {
  to = azurerm_storage_queue.equislice-sa-queue
  id = "${local.storage_id}/queueServices/default/queues/panorama-slice"
}
import {
  to = azurerm_storage_queue.poison
  id = "${local.storage_id}/queueServices/default/queues/panorama-slice-poison"
}
import {
  to = azurerm_storage_table.equislice-sa-table
  id = "https://${local.storage_name}.table.core.windows.net/Tables('Panorama')"
}
import {
  to = azurerm_service_plan.backend-sp
  id = "${local.resource_group_id}/providers/Microsoft.Web/serverFarms/equislice-api-japaneast-plan"
}
import {
  to = azurerm_linux_web_app.backend
  id = local.backend_id
}
import {
  to = azurerm_service_plan.function
  id = "${local.resource_group_id}/providers/Microsoft.Web/serverFarms/ASP-rgrenatodsantosjr91317-af55"
}
import {
  to = azurerm_application_insights.function
  id = "${local.resource_group_id}/providers/Microsoft.Insights/components/${local.function_name}"
}
import {
  to = azapi_resource.function
  id = "${local.function_id}?api-version=2024-04-01"
}
import {
  to = azurerm_role_assignment.github_backend
  id = "${local.backend_id}/providers/Microsoft.Authorization/roleAssignments/98989f27-3011-4daa-a8ab-ece1bb3c97aa"
}
import {
  to = azurerm_role_assignment.github_function
  id = "${local.function_id}/providers/Microsoft.Authorization/roleAssignments/80df005d-1dfc-4945-8054-95ce694a1b3d"
}
