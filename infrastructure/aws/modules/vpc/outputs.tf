output "vpc_id" {
  description = "ID of the VPC."
  value       = aws_vpc.this.id
}

output "vpc_cidr_block" {
  description = "CIDR block of the VPC."
  value       = aws_vpc.this.cidr_block
}

output "public_subnet_ids" {
  description = "IDs of the public subnets."
  value       = [for az in var.azs : aws_subnet.public[az].id]
}

output "private_subnet_ids" {
  description = "IDs of the private subnets."
  value       = [for az in var.azs : aws_subnet.private[az].id]
}

output "nat_public_ips" {
  description = "Public IPs of the NAT gateways."
  value       = [for eip in aws_eip.nat : eip.public_ip]
}

output "public_route_table_id" {
  description = "ID of the public route table."
  value       = aws_route_table.public.id
}

output "private_route_table_ids" {
  description = "IDs of the private route tables keyed by AZ."
  value       = { for az, rt in aws_route_table.private : az => rt.id }
}
