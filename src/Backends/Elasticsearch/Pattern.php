<?php
/**
 * Statusengine UI
 * Copyright (C) 2017  Daniel Ziegler
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

namespace Statusengine\Backend\Elasticsearch;


use Statusengine\Config;
use Statusengine\ValueObjects\ServicePerfdataQueryOptions;

class Pattern {

    /**
     * @var string
     */
    private $pattern;

    /**
     * @var string
     */
    private $index;

    const ONE_DAY = 60 * 60 * 24;

    const ONE_WEEK = 60 * 60 * 24 * 7;

    //more or less
    const ONE_MONTH = 60 * 60 * 24 * 28;

    public function __construct(Config $Config) {
        $this->pattern = $Config->getElasticsearchPattern();
        $this->index = $Config->getElasticsearchIndex();
    }

    /**
     * @param ServicePerfdataQueryOptions $ServicePerfdataQueryOptions
     * @return array
     */
    public function getPossibleIndices(ServicePerfdataQueryOptions $ServicePerfdataQueryOptions) {
        //Get shards where timestamp is > $end and < $start
        $end = $ServicePerfdataQueryOptions->getEnd(); //right side of graph
        $start = $ServicePerfdataQueryOptions->getStart(); //left side of graph

        if ($end > $start) {
            $end = time() - 60 * 60 * 2.5;
            $start = time();
        }

        if ($this->pattern === 'none') {
            return [$this->index];
        }

        switch ($this->pattern) {
            case 'daily':
                return $this->getIndicesForDayPattern($start, $end);
            case 'weekly':
                return $this->getIndicesForWeekPattern($start, $end);
            case 'monthly':
                return $this->getIndicesForMonthPattern($start, $end);
            default:
                return [$this->index];
        }
    }

    /**
     * @param int $start
     * @param int $end
     * @return array
     */
    private function getIndicesForDayPattern($start, $end) {
        if (($start - $end) > self::ONE_DAY) {
            $period = new \DatePeriod(
                new \DateTime(date('Y-m-d', $end)),
                new \DateInterval('P1D'),
                new \DateTime(date('Y-m-d', $start))
            );

            $days = [];
            /**
             * @var \DateTime $day
             */
            foreach ($period as $day) {
                $days[] = sprintf('%s%s', $this->index, $day->format('Y.m.d'));
            }
            $days[] = sprintf('%s%s', $this->index, date('Y.m.d'));
            return array_unique($days);
        }

        if (date('Y.m.d', $end) !== date('Y.m.d', $start)) {
            return [
                sprintf(
                    '%s%s',
                    $this->index, date('Y.m.d', $end)
                ),
                sprintf(
                    '%s%s',
                    $this->index, date('Y.m.d', $start)
                )
            ];
        }

        return [
            sprintf(
                '%s%s',
                $this->index, date('Y.m.d', $end)
            )
        ];
    }

    /**
     * @param int $start
     * @param int $end
     * @return array
     */
    private function getIndicesForWeekPattern($start, $end) {
        if (($start - $end) > self::ONE_WEEK) {
            $period = new \DatePeriod(
                new \DateTime(date('Y-m-d', $end)),
                new \DateInterval('P1W'),
                new \DateTime(date('Y-m-d', $start))
            );

            $weeks = [];
            /**
             * @var \DateTime $week
             */
            foreach ($period as $week) {
                $weeks[] = sprintf('%s%s', $this->index, $week->format('o.W'));
            }
            $weeks[] = sprintf('%s%s', $this->index, date('o.W'));
            return array_unique($weeks);
        }

        if (date('o.W', $end) !== date('o.W', $start)) {
            return [
                sprintf(
                    '%s%s',
                    $this->index, date('o.W', $end)
                ),
                sprintf(
                    '%s%s',
                    $this->index, date('o.W', $start)
                )
            ];
        }

        return [
            sprintf(
                '%s%s',
                $this->index, date('o.W', $end)
            )
        ];
    }

    /**
     * @param int $start
     * @param int $end
     * @return array
     */
    private function getIndicesForMonthPattern($start, $end) {
        if (($start - $end) > self::ONE_MONTH) {
            $period = new \DatePeriod(
                new \DateTime(date('Y-m-d', $end)),
                new \DateInterval('P1M'),
                new \DateTime(date('Y-m-d', $start))
            );

            $months = [];
            /**
             * @var \DateTime $week
             */
            foreach ($period as $month) {
                $months[] = sprintf('%s%s', $this->index, $month->format('Y.m'));
            }
            $months[] = sprintf('%s%s', $this->index, date('Y.m'));
            return array_unique($months);
        }

        if (date('Y.m', $end) !== date('Y.m', $start)) {
            return [
                sprintf(
                    '%s%s',
                    $this->index, date('Y.m', $end)
                ),
                sprintf(
                    '%s%s',
                    $this->index, date('Y.m', $start)
                )
            ];
        }

        return [
            sprintf(
                '%s%s',
                $this->index, date('Y.m', $end)
            )
        ];
    }

}