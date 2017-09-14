<?php
/**
 * Statusengine UI
 * Copyright (C) 2016-2017  Daniel Ziegler

 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

namespace Statusengine\Loader;


use Statusengine\Backend\StorageBackend;

interface DashboardLoaderInterface {

    /**
     * DashboardLoaderInterface constructor.
     * @param StorageBackend $StorageBackend
     */
    public function __construct(StorageBackend $StorageBackend);


    /**
     * @return int
     */
    public function getNumberOfMonitoredHosts();

    /**
     * @return int
     */
    public function getNumberOfMonitoredServices();

    /**
     * @param array $states
     * @return array
     */
    public function getHostOverview($states = [0, 1, 2]);

    /**
     * @param array $states
     * @return array
     */
    public function getServiceOverview($states = [0, 1, 2, 3]);

    /**
     * @return int
     */
    public function getNumberOfServiceProblems();

    /**
     * @return int
     */
    public function getNumberOfHostAcknowledgements();

    /**
     * @return int
     */
    public function getNumberOfServiceAcknowledgements();

    /**
     * @return int
     */
    public function getNummerOfScheduledHostDowntimes();

    /**
     * @return int
     */
    public function getNummerOfScheduledServiceDowntimes();

}