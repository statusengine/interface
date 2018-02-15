<?php
/**
 * Statusengine UI
 * Copyright (C) 2016-2018  Daniel Ziegler
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


use Statusengine\Backend\StorageBackend;
use Statusengine\Loader\ServiceDowntimeLoaderInterface;
use Statusengine\ValueObjects\ServiceDowntimeQueryOptions;

class ServiceDowntimeLoader implements ServiceDowntimeLoaderInterface {

    /**
     * @var \Statusengine\Backend\Crate\Crate
     */
    private $Backend;

    /**
     * ServiceDowntimeLoader constructor.
     * @param StorageBackend $StorageBackend
     */
    public function __construct(StorageBackend $StorageBackend) {
        $this->Backend = $StorageBackend->getBackend();
    }

    public function getServicedowntime(ServiceDowntimeQueryOptions $ServiceDowntimeQueryOptions) {
        $fields = [
            'SDH.internal_downtime_id',
            'SDH.scheduled_start_time',
            'SDH.scheduled_end_time',
            'SDH.author_name',
            'SDH.comment_data',
            'SDH.is_fixed',
            'SDH.node_name',

            'SSD.hostname',
            'SSD.service_description',
            'SSD.actual_start_time'

        ];
        $sql = "SELECT %s FROM statusengine_service_downtimehistory AS SDH
             INNER JOIN statusengine_service_scheduleddowntimes AS SSD ON
                 SDH.internal_downtime_id = SSD.internal_downtime_id
              WHERE SSD.actual_start_time > 0
              AND SSD.hostname=? AND SSD.service_description=?
              ORDER BY SDH.scheduled_start_time DESC
             LIMIT 1";

        $sql = sprintf($sql, implode(', ', $fields));

        $query = $this->Backend->prepare($sql);
        $query->bindValue(1, $ServiceDowntimeQueryOptions->getHostname());
        $query->bindValue(2, $ServiceDowntimeQueryOptions->getServiceDescription());
        $result = $this->Backend->fetchAll($query);
        if (!empty($result)) {
            return $result;
        }
        return [];
    }
}
