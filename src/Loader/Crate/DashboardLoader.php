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

namespace Statusengine\Loader\Crate;


use Statusengine\Backend\Crate\Crate;
use Statusengine\Backend\StorageBackend;
use Statusengine\Loader\DashboardLoaderInterface;

class DashboardLoader implements DashboardLoaderInterface {

    /**
     * @var StorageBackend
     */
    private $StorageBackend;

    /**
     * @var \Statusengine\Backend\Crate\Crate
     */
    private $Backend;

    /**
     * DashboardLoader constructor.
     * @param StorageBackend $StorageBackend
     */
    public function __construct(StorageBackend $StorageBackend) {
        $this->Backend = $StorageBackend->getBackend();
    }


    /**
     * @return int
     */
    public function getNumberOfMonitoredHosts() {
        $query = $this->Backend->prepare('SELECT COUNT(*) as count FROM statusengine_hoststatus');
        $result = $this->Backend->fetchAll($query);
        return (int)$result[0]['count'];
    }

    /**
     * @return int
     */
    public function getNumberOfMonitoredServices() {
        $query = $this->Backend->prepare('SELECT COUNT(*) as count FROM statusengine_servicestatus');
        $result = $this->Backend->fetchAll($query);
        return (int)$result[0]['count'];
    }

    /**
     * @param array $states
     * @return array
     */
    public function getHostOverview($states = [0, 1, 2]) {
        $result = [
            0 => 0, //up
            1 => 0, //down
            2 => 0, //unreachable
        ];
        $baseQuery = 'SELECT current_state, COUNT(current_state) AS count FROM statusengine_hoststatus %s GROUP BY current_state';
        if (sizeof($states) < 3) {
            $baseQuery = sprintf(
                $baseQuery,
                sprintf('WHERE current_state IN (%s)', implode(',', $states))
            );
        } else {
            $baseQuery = sprintf($baseQuery, '');
        }

        $query = $this->Backend->prepare($baseQuery);
        foreach ($this->Backend->fetchAll($query) as $currentState) {
            $result[$currentState['current_state']] = (int)$currentState['count'];
        }

        return $result;
    }

    /**
     * @param array $states
     * @return array
     */
    public function getServiceOverview($states = [0, 1, 2, 3]) {
        $result = [
            0 => 0, //ok
            1 => 0, //warning
            2 => 0, //critical
            3 => 0, //unknown
        ];
        $baseQuery = 'SELECT current_state, COUNT(current_state) AS count FROM statusengine_servicestatus %s GROUP BY current_state';
        if (sizeof($states) < 4) {
            $baseQuery = sprintf(
                $baseQuery,
                sprintf('WHERE current_state IN (%s)', implode(',', $states))
            );
        } else {
            $baseQuery = sprintf($baseQuery, '');
        }

        $query = $this->Backend->prepare($baseQuery);
        foreach ($this->Backend->fetchAll($query) as $currentState) {
            $result[$currentState['current_state']] = (int)$currentState['count'];
        }

        return $result;
    }

    /**
     * @return int
     */
    public function getNumberOfServiceProblems() {
        $query = $this->Backend->prepare('
          SELECT COUNT(*) AS count FROM statusengine_servicestatus 
          WHERE current_state > 0
          AND problem_has_been_acknowledged = 0
          AND scheduled_downtime_depth = 0
        ');
        $result = $this->Backend->fetchAll($query);
        return (int)$result[0]['count'];
    }

    /**
     * @return int
     */
    public function getNumberOfHostAcknowledgements(){
        $baseQuery = 'SELECT COUNT(*) AS count FROM statusengine_hoststatus WHERE problem_has_been_acknowledged=true';
        $query = $this->Backend->prepare($baseQuery);
        $result = $this->Backend->fetchAll($query);

        if(isset($result[0]['count'])){
            return (int)$result[0]['count'];
        }

        return 0;
    }

    /**
     * @return int
     */
    public function getNumberOfServiceAcknowledgements(){
        $baseQuery = 'SELECT COUNT(*) AS count FROM statusengine_servicestatus WHERE problem_has_been_acknowledged=true';
        $query = $this->Backend->prepare($baseQuery);
        $result = $this->Backend->fetchAll($query);

        if(isset($result[0]['count'])){
            return (int)$result[0]['count'];
        }

        return 0;
    }

    /**
     * @return int
     */
    public function getNummerOfScheduledHostDowntimes(){
        $baseQuery = 'SELECT COUNT(*) AS count FROM statusengine_hoststatus WHERE scheduled_downtime_depth > 0';
        $query = $this->Backend->prepare($baseQuery);
        $result = $this->Backend->fetchAll($query);

        if(isset($result[0]['count'])){
            return (int)$result[0]['count'];
        }

        return 0;
    }

    /**
     * @return int
     */
    public function getNummerOfScheduledServiceDowntimes(){
        $baseQuery = 'SELECT COUNT(*) AS count FROM statusengine_servicestatus WHERE scheduled_downtime_depth > 0';
        $query = $this->Backend->prepare($baseQuery);
        $result = $this->Backend->fetchAll($query);

        if(isset($result[0]['count'])){
            return (int)$result[0]['count'];
        }

        return 0;
    }


}